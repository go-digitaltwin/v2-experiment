package deltagen

import (
	_ "embed"
	"fmt"
	"io"
	"strings"
	"text/template"
)

//go:embed delta.go.tmpl
var deltaTemplate string

// templateData carries the resolved values for template rendering.
type templateData struct {
	Command    string // full invocation for the generated header
	PkgName    string
	TypeName   string
	DeltaName  string
	Keys       []FieldInfo
	Properties []FieldInfo
	KeyLinks   string // e.g. "[T.IMEI]" or "[T.IMSI], [T.APN]"
	KeyParams  string // e.g. "imei string" or "tenantID string, name string"
	KeyAssigns string // e.g. "imei: imei" or "tenantID: tenantID, name: name"
	Imports    []string
	Apply      bool
}

func newTemplateData(ti TypeInfo, cfg Config) templateData {
	keys := ti.Keys()
	keyLinks := make([]string, len(keys))
	params := make([]string, len(keys))
	assigns := make([]string, len(keys))
	for i, k := range keys {
		keyLinks[i] = fmt.Sprintf("[%s.%s]", ti.TypeName, k.Name)
		params[i] = fmt.Sprintf("%s %s", lowerFirst(k.Name), k.TypeStr)
		assigns[i] = fmt.Sprintf("%s: %s", lowerFirst(k.Name), lowerFirst(k.Name))
	}

	return templateData{
		Command:    cfg.Command,
		PkgName:    ti.PkgName,
		TypeName:   ti.TypeName,
		DeltaName:  ti.TypeName + "Delta",
		Keys:       keys,
		Properties: ti.Properties(),
		KeyLinks:   strings.Join(keyLinks, ", "),
		KeyParams:  strings.Join(params, ", "),
		KeyAssigns: strings.Join(assigns, ", "),
		Imports:    collectImports(ti),
		Apply:      cfg.Apply,
	}
}

var funcMap = template.FuncMap{
	"wrap":       wrapComment,
	"lowerFirst": lowerFirst,
	"keyLink": func(typeName, fieldName string) string {
		return fmt.Sprintf("[%s.%s]", typeName, fieldName)
	},
}

var tmpl = template.Must(template.New("delta").Funcs(funcMap).Parse(deltaTemplate))

func emit(w io.Writer, ti TypeInfo, cfg Config) {
	data := newTemplateData(ti, cfg)
	if err := tmpl.Execute(w, data); err != nil {
		panic(fmt.Sprintf("template execution failed: %v", err))
	}
	fmt.Fprintln(w) // trailing newline
}
