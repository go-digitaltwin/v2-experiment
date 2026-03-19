package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-digitaltwin/v2-experiment/internal/deltagen"
)

func main() {
	typeName := flag.String("type", "", "target struct name (required)")
	key := flag.String("key", "", "comma-separated primary key field names (required)")
	output := flag.String("output", "", "output file path (optional; defaults to ${GOFILE%.go}_delta.go)")
	dir := flag.String("dir", ".", "package directory (optional; defaults to current directory)")
	apply := flag.Bool("apply", false, "generate Apply method")
	flag.Parse()

	if *typeName == "" || *key == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg := deltagen.Config{
		TypeName: *typeName,
		Keys:     strings.Split(*key, ","),
		Dir:      *dir,
		Apply:    *apply,
		Command:  "gen-delta-builder " + strings.Join(os.Args[1:], " "),
	}

	src, err := deltagen.Generate(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gen-delta-builder: %v\n", err)
		os.Exit(1)
	}

	outPath := outputPath(*output, *typeName)
	if err := os.WriteFile(outPath, src, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "gen-delta-builder: writing %s: %v\n", outPath, err)
		os.Exit(1)
	}
}

func outputPath(explicit, typeName string) string {
	if explicit != "" {
		return explicit
	}
	if gofile := os.Getenv("GOFILE"); gofile != "" {
		return strings.TrimSuffix(gofile, filepath.Ext(gofile)) + "_delta.go"
	}
	return strings.ToLower(typeName) + "_delta.go"
}
