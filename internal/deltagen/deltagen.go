package deltagen

import (
	"bytes"
	"fmt"
	"go/format"
	"go/types"
	"maps"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

// FieldKind classifies a struct field as either part of the primary key or a
// mutable property.
type FieldKind int

const (
	FieldProperty FieldKind = iota
	FieldKey
)

// FieldInfo describes a single struct field.
type FieldInfo struct {
	Name    string
	Kind    FieldKind
	TypeStr string // qualified type string (e.g. "netip.Addr", "[]string")
	PkgPath string // import path for non-builtin types ("net/netip"), empty otherwise
}

// TypeInfo carries the parsed metadata for the target struct.
type TypeInfo struct {
	PkgName  string
	TypeName string
	Fields   []FieldInfo
}

// Keys returns only the primary-key fields.
func (ti TypeInfo) Keys() []FieldInfo {
	var out []FieldInfo
	for _, f := range ti.Fields {
		if f.Kind == FieldKey {
			out = append(out, f)
		}
	}
	return out
}

// Properties returns only the non-key fields.
func (ti TypeInfo) Properties() []FieldInfo {
	var out []FieldInfo
	for _, f := range ti.Fields {
		if f.Kind == FieldProperty {
			out = append(out, f)
		}
	}
	return out
}

// Config carries the inputs for the generator.
type Config struct {
	TypeName string   // -type value
	Keys     []string // -key value, split by comma
	Dir      string   // package directory (from working dir for now)
	Apply    bool     // -apply flag
	Command  string   // full invocation for the generated header
}

// Generate loads the target struct from Dir, classifies its fields, and
// returns the formatted generated source.
func Generate(cfg Config) ([]byte, error) {
	ti, err := loadType(cfg)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := emit(&buf, ti, cfg); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return format.Source(buf.Bytes())
}

func loadType(cfg Config) (TypeInfo, error) {
	pkgCfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax,
		Dir:  cfg.Dir,
	}
	pkgs, err := packages.Load(pkgCfg, ".")
	if err != nil {
		return TypeInfo{}, fmt.Errorf("loading package: %w", err)
	}
	if len(pkgs) == 0 {
		return TypeInfo{}, fmt.Errorf("no packages found in %s", cfg.Dir)
	}
	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		return TypeInfo{}, fmt.Errorf("package errors: %v", pkg.Errors)
	}

	obj := pkg.Types.Scope().Lookup(cfg.TypeName)
	if obj == nil {
		return TypeInfo{}, fmt.Errorf("type %s not found in package %s", cfg.TypeName, pkg.Name)
	}
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return TypeInfo{}, fmt.Errorf("%s is not a named type", cfg.TypeName)
	}
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		return TypeInfo{}, fmt.Errorf("%s is not a struct type", cfg.TypeName)
	}

	keySet := make(map[string]bool, len(cfg.Keys))
	for _, k := range cfg.Keys {
		keySet[k] = true
	}

	ti := TypeInfo{
		PkgName:  pkg.Name,
		TypeName: cfg.TypeName,
	}
	for field := range st.Fields() {
		if !field.Exported() {
			continue
		}
		kind := FieldProperty
		if keySet[field.Name()] {
			kind = FieldKey
		}
		typeStr, pkgPath := qualifiedType(field.Type(), pkg.Types)
		ti.Fields = append(ti.Fields, FieldInfo{
			Name:    field.Name(),
			Kind:    kind,
			TypeStr: typeStr,
			PkgPath: pkgPath,
		})
	}

	// Validate all declared keys were found.
	for _, k := range cfg.Keys {
		if !slices.ContainsFunc(ti.Fields, func(f FieldInfo) bool {
			return f.Name == k && f.Kind == FieldKey
		}) {
			return TypeInfo{}, fmt.Errorf("key field %q not found in struct %s", k, cfg.TypeName)
		}
	}

	return ti, nil
}

// qualifiedType returns the type string suitable for generated code and the
// import path of the package if the type requires one.
func qualifiedType(t types.Type, localPkg *types.Package) (typeStr string, pkgPath string) {
	switch t := t.(type) {
	case *types.Named:
		obj := t.Obj()
		if obj.Pkg() == nil || obj.Pkg() == localPkg {
			return obj.Name(), ""
		}
		return obj.Pkg().Name() + "." + obj.Name(), obj.Pkg().Path()
	case *types.Array:
		elemStr, elemPkg := qualifiedType(t.Elem(), localPkg)
		return fmt.Sprintf("[%d]%s", t.Len(), elemStr), elemPkg
	case *types.Slice:
		elemStr, elemPkg := qualifiedType(t.Elem(), localPkg)
		return "[]" + elemStr, elemPkg
	case *types.Map:
		keyStr, keyPkg := qualifiedType(t.Key(), localPkg)
		valStr, valPkg := qualifiedType(t.Elem(), localPkg)
		pkg := keyPkg
		if pkg == "" {
			pkg = valPkg
		}
		return "map[" + keyStr + "]" + valStr, pkg
	case *types.Pointer:
		elemStr, elemPkg := qualifiedType(t.Elem(), localPkg)
		return "*" + elemStr, elemPkg
	case *types.Basic:
		return t.Name(), ""
	default:
		return types.TypeString(t, nil), ""
	}
}

// collectImports returns the external import paths referenced by any field
// (keys and properties). Order is non-deterministic; format.Source sorts the
// import block in the generated output.
func collectImports(ti TypeInfo) []string {
	imports := make(map[string]struct{})
	for _, f := range ti.Fields {
		if f.PkgPath != "" {
			imports[f.PkgPath] = struct{}{}
		}
	}
	return slices.Collect(maps.Keys(imports))
}

// lowerFirst converts the first character of s to lower case.
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	// Handle all-uppercase abbreviations like "IMEI" -> "imei"
	if len(runes) > 1 && unicode.IsUpper(runes[1]) {
		return strings.ToLower(s)
	}
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
