package deltagen_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/go-digitaltwin/v2-experiment/cmd/gen-delta-builder/internal/deltagen"
)

var update = flag.Bool("update", false, "update golden files")

func TestGenerate(t *testing.T) {
	tests := []struct {
		name   string
		dir    string
		typ    string
		keys   []string
		golden string
	}{
		{
			name:   "simple",
			dir:    "testdata/simple",
			typ:    "Device",
			keys:   []string{"ID"},
			golden: "testdata/simple/simple.golden",
		},
		{
			name:   "composite",
			dir:    "testdata/composite",
			typ:    "Connection",
			keys:   []string{"TenantID", "Name"},
			golden: "testdata/composite/composite.golden",
		},
		{
			name:   "keyonly",
			dir:    "testdata/keyonly",
			typ:    "Label",
			keys:   []string{"Name"},
			golden: "testdata/keyonly/keyonly.golden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := deltagen.Config{
				TypeName: tt.typ,
				Keys:     tt.keys,
				Dir:      tt.dir,
			}
			got, err := deltagen.Generate(cfg)
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}

			goldenPath, _ := filepath.Abs(tt.golden)
			if *update {
				if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
					t.Fatalf("updating golden file: %v", err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("reading golden file: %v", err)
			}

			if diff := cmp.Diff(string(want), string(got)); diff != "" {
				actualPath := tt.golden[:len(tt.golden)-len(".golden")] + ".actual"
				_ = os.WriteFile(actualPath, got, 0o644)
				t.Errorf("output mismatch (-want +got, wrote %s):\n%s", actualPath, diff)
			}
		})
	}
}
