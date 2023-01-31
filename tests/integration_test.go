package tests

import (
	"github.com/lets-dev-it-out/go-jsonschema/pkg/generator"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var basicConfig = generator.Config{
	SchemaMappings:     []generator.SchemaMapping{},
	DefaultPackageName: "github.com/example/test",
	DefaultOutputName:  "-",
	ResolveExtensions:  []string{".json", ".yaml"},
	Warner: func(message string) {
		log.Printf("[from warner] %s", message)
	},
}

func TestCore(t *testing.T) {
	testExamples(t, basicConfig, "./data/core")
}

func TestValidation(t *testing.T) {
	testExamples(t, basicConfig, "./data/validation")
}

func TestMiscWithDefaults(t *testing.T) {
	testExamples(t, basicConfig, "./data/miscWithDefaults")
}

func TestCrossPackage(t *testing.T) {
	cfg := basicConfig
	cfg.SchemaMappings = []generator.SchemaMapping{
		{
			SchemaID:    "https://example.com/schema",
			PackageName: "github.com/example/schema",
			OutputName:  "schema.go",
		},
		{
			SchemaID:    "https://example.com/other",
			PackageName: "github.com/example/other",
			OutputName:  "other.go",
		},
	}
	testExampleFile(t, cfg, "./data/crossPackage/schema.json")
}

func TestCrossPackageNoOutput(t *testing.T) {
	cfg := basicConfig
	cfg.SchemaMappings = []generator.SchemaMapping{
		{
			SchemaID:    "https://example.com/schema",
			PackageName: "github.com/example/schema",
			OutputName:  "schema.go",
		},
		{
			SchemaID:    "https://example.com/other",
			PackageName: "github.com/example/other",
		},
	}
	testExampleFile(t, cfg, "./data/crossPackageNoOutput/schema.json")
}

func TestCapitalization(t *testing.T) {
	cfg := basicConfig
	cfg.Capitalizations = []string{"ID", "URL", "HtMl"}
	testExampleFile(t, cfg, "./data/misc/capitalization.json")
}

func TestBooleanAsSchema(t *testing.T) {
	cfg := basicConfig
	testExampleFile(t, cfg, "./data/misc/boolean-as-schema.json")
}

func testExamples(t *testing.T, cfg generator.Config, dataDir string) {
	fileInfos, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, file := range fileInfos {
		if strings.HasSuffix(file.Name(), ".json") {
			fileName := filepath.Join(dataDir, file.Name())
			if strings.HasSuffix(file.Name(), ".FAIL.json") {
				testFailingExampleFile(t, cfg, fileName)
			} else {
				testExampleFile(t, cfg, fileName)
			}
		}
	}
}

func testExampleFile(t *testing.T, cfg generator.Config, fileName string) {
	t.Run(titleFromFileName(fileName), func(t *testing.T) {
		generator, err := generator.New(cfg)
		if err != nil {
			t.Fatal(err)
		}

		if err := generator.DoFile(fileName); err != nil {
			t.Fatal(err)
		}

		if len(generator.Sources()) == 0 {
			t.Fatal("Expected sources to contain something")
		}

		for outputName, source := range generator.Sources() {
			if outputName == "-" {
				outputName = strings.TrimSuffix(filepath.Base(fileName), ".json") + ".go"
			}
			outputName += ".output"

			goldenFileName := filepath.Join(filepath.Dir(fileName), outputName)
			t.Logf("Using golden data in %s", mustAbs(goldenFileName))

			goldenData, err := os.ReadFile(goldenFileName)
			if err != nil {
				if !os.IsNotExist(err) {
					t.Fatal(err)
				}
				goldenData = source
				t.Log("File does not exist; creating it")
				if err = os.WriteFile(goldenFileName, goldenData, 0655); err != nil {
					t.Fatal(err)
				}
			}

			require.Equal(t, string(goldenData), string(source))
		}
	})
}

func testFailingExampleFile(t *testing.T, cfg generator.Config, fileName string) {
	t.Run(titleFromFileName(fileName), func(t *testing.T) {
		generator, err := generator.New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if err := generator.DoFile(fileName); err == nil {
			t.Fatal("Expected test to fail")
		}
	})
}

func titleFromFileName(fileName string) string {
	relative := mustRel(mustAbs("./data"), mustAbs(fileName))
	return strings.TrimSuffix(relative, ".json")
}

func mustRel(base, s string) string {
	result, err := filepath.Rel(base, s)
	if err != nil {
		panic(err)
	}
	return result
}

func mustAbs(s string) string {
	result, err := filepath.Abs(s)
	if err != nil {
		panic(err)
	}
	return result
}
