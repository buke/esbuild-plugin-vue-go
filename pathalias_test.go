// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package vueplugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
)

// TestParseTsconfigPathAliasWithTsconfigRaw tests parsing path aliases from raw tsconfig JSON.
func TestParseTsconfigPathAliasWithTsconfigRaw(t *testing.T) {
	buildOptions := &api.BuildOptions{
		TsconfigRaw:   `{"compilerOptions":{"paths":{"@/*":["./src/*"],"@utils/*":["./src/utils/*"]}}}`,
		AbsWorkingDir: "/test/project",
	}

	pathAlias, err := parseTsconfigPathAlias(buildOptions)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedAlias := filepath.Join("/test/project", "./src/*")
	if pathAlias["@/*"] != expectedAlias {
		t.Errorf("Expected alias '@/*' to be '%s', got '%s'", expectedAlias, pathAlias["@/*"])
	}

	expectedUtilsAlias := filepath.Join("/test/project", "./src/utils/*")
	if pathAlias["@utils/*"] != expectedUtilsAlias {
		t.Errorf("Expected alias '@utils/*' to be '%s', got '%s'", expectedUtilsAlias, pathAlias["@utils/*"])
	}
}

// TestParseTsconfigPathAliasWithInvalidRawJSON tests parsing with invalid raw JSON.
func TestParseTsconfigPathAliasWithInvalidRawJSON(t *testing.T) {
	buildOptions := &api.BuildOptions{
		TsconfigRaw: `{invalid json}`,
	}

	_, err := parseTsconfigPathAlias(buildOptions)
	if err == nil {
		t.Errorf("Expected error with invalid JSON, got nil")
	}
}

// TestParseTsconfigPathAliasWithTsconfigRawNoAbsWorkingDir tests raw JSON without AbsWorkingDir.
func TestParseTsconfigPathAliasWithTsconfigRawNoAbsWorkingDir(t *testing.T) {
	buildOptions := &api.BuildOptions{
		TsconfigRaw: `{"compilerOptions":{"paths":{"@/*":["./src/*"]}}}`,
		// AbsWorkingDir is empty, should use executable path
	}

	pathAlias, err := parseTsconfigPathAlias(buildOptions)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should have some path set (using executable path)
	if len(pathAlias) == 0 {
		t.Errorf("Expected some path aliases to be parsed")
	}
}

// TestParseTsconfigPathAliasWithTsconfigFile tests parsing from tsconfig file.
func TestParseTsconfigPathAliasWithTsconfigFile(t *testing.T) {
	// Create a temporary tsconfig file
	tmpDir := t.TempDir()
	tsconfigPath := filepath.Join(tmpDir, "tsconfig.json")
	tsconfigContent := `{
        "compilerOptions": {
            "paths": {
                "@/*": ["./src/*"],
                "@components/*": ["./src/components/*"]
            }
        }
    }`

	err := os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test tsconfig file: %v", err)
	}

	buildOptions := &api.BuildOptions{
		Tsconfig: tsconfigPath,
	}

	pathAlias, err := parseTsconfigPathAlias(buildOptions)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedAlias := filepath.Join(tmpDir, "./src/*")
	if pathAlias["@/*"] != expectedAlias {
		t.Errorf("Expected alias '@/*' to be '%s', got '%s'", expectedAlias, pathAlias["@/*"])
	}

	expectedComponentsAlias := filepath.Join(tmpDir, "./src/components/*")
	if pathAlias["@components/*"] != expectedComponentsAlias {
		t.Errorf("Expected alias '@components/*' to be '%s', got '%s'", expectedComponentsAlias, pathAlias["@components/*"])
	}
}

// TestParseTsconfigPathAliasWithNonexistentFile tests parsing with nonexistent tsconfig file.
func TestParseTsconfigPathAliasWithNonexistentFile(t *testing.T) {
	buildOptions := &api.BuildOptions{
		Tsconfig: "/nonexistent/tsconfig.json",
	}

	_, err := parseTsconfigPathAlias(buildOptions)
	if err == nil {
		t.Errorf("Expected error with nonexistent file, got nil")
	}
}

// TestParseTsconfigPathAliasWithInvalidFileJSON tests parsing with invalid JSON in file.
func TestParseTsconfigPathAliasWithInvalidFileJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tsconfigPath := filepath.Join(tmpDir, "tsconfig.json")
	invalidContent := `{invalid json}`

	err := os.WriteFile(tsconfigPath, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test tsconfig file: %v", err)
	}

	buildOptions := &api.BuildOptions{
		Tsconfig: tsconfigPath,
	}

	_, err = parseTsconfigPathAlias(buildOptions)
	if err == nil {
		t.Errorf("Expected error with invalid JSON file, got nil")
	}
}

// TestParseTsconfigPathAliasEmpty tests parsing with empty build options.
func TestParseTsconfigPathAliasEmpty(t *testing.T) {
	buildOptions := &api.BuildOptions{}

	pathAlias, err := parseTsconfigPathAlias(buildOptions)
	if err != nil {
		t.Errorf("Expected no error with empty options, got: %v", err)
	}

	if len(pathAlias) != 0 {
		t.Errorf("Expected empty path alias map, got %d entries", len(pathAlias))
	}
}

// TestApplyPathAliasWithWildcard tests path alias application with wildcard.
func TestApplyPathAliasWithWildcard(t *testing.T) {
	pathAlias := map[string]string{
		"@/*":        "/src/*",
		"@utils/*":   "/src/utils/*",
		"@exact":     "/src/exact",
		"@special/*": "/very/long/path/to/special/*",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"@/components/Button.vue", "/src/components/Button.vue"},
		{"@utils/helper.ts", "/src/utils/helper.ts"},
		{"@exact", "/src/exact"},
		{"@special/deep/nested/file.ts", "/very/long/path/to/special/deep/nested/file.ts"},
		{"normal/path", "normal/path"},     // No alias match
		{"@unknown/path", "@unknown/path"}, // No alias match
	}

	for _, test := range tests {
		result := applyPathAlias(pathAlias, test.input)
		if result != test.expected {
			t.Errorf("applyPathAlias(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

// TestApplyPathAliasWithoutWildcard tests path alias application without wildcard.
func TestApplyPathAliasWithoutWildcard(t *testing.T) {
	pathAlias := map[string]string{
		"@exact": "/src/exact",
		"@lib":   "/node_modules/lib",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"@exact", "/src/exact"},
		{"@lib", "/node_modules/lib"},
		{"@exact/sub", "@exact/sub"}, // No match for exact alias with sub path
		{"normal", "normal"},         // No alias match
	}

	for _, test := range tests {
		result := applyPathAlias(pathAlias, test.input)
		if result != test.expected {
			t.Errorf("applyPathAlias(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

// TestApplyPathAliasEmptyMap tests path alias application with empty alias map.
func TestApplyPathAliasEmptyMap(t *testing.T) {
	pathAlias := map[string]string{}
	input := "@/components/Button.vue"

	result := applyPathAlias(pathAlias, input)
	if result != input {
		t.Errorf("Expected unchanged path with empty alias map, got: %s", result)
	}
}

// TestParseTsconfigPathAliasWithComplexPaths tests parsing with complex path configurations.
func TestParseTsconfigPathAliasWithComplexPaths(t *testing.T) {
	buildOptions := &api.BuildOptions{
		TsconfigRaw: `{
            "compilerOptions": {
                "paths": {
                    "@/*": ["./src/*"],
                    "#/*": ["./types/*"],
                    "utils": ["./src/utils/index.ts"],
                    "empty": [],
                    "multiple": ["./first/*", "./second/*"],
                    "invalid": "not_an_array"
                }
            }
        }`,
		AbsWorkingDir: "/project",
	}

	pathAlias, err := parseTsconfigPathAlias(buildOptions)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should parse valid entries
	expectedSrc := filepath.Join("/project", "./src/*")
	if pathAlias["@/*"] != expectedSrc {
		t.Errorf("Expected '@/*' alias, got: %s", pathAlias["@/*"])
	}

	expectedTypes := filepath.Join("/project", "./types/*")
	if pathAlias["#/*"] != expectedTypes {
		t.Errorf("Expected '#/*' alias, got: %s", pathAlias["#/*"])
	}

	// Should take first entry for multiple paths
	expectedUtils := filepath.Join("/project", "./src/utils/index.ts")
	if pathAlias["utils"] != expectedUtils {
		t.Errorf("Expected 'utils' alias, got: %s", pathAlias["utils"])
	}

	expectedMultiple := filepath.Join("/project", "./first/*")
	if pathAlias["multiple"] != expectedMultiple {
		t.Errorf("Expected 'multiple' alias to use first path, got: %s", pathAlias["multiple"])
	}

	// Should skip empty arrays and invalid entries
	if _, exists := pathAlias["empty"]; exists {
		t.Errorf("Expected 'empty' alias to be skipped")
	}
	if _, exists := pathAlias["invalid"]; exists {
		t.Errorf("Expected 'invalid' alias to be skipped")
	}
}

// TestParseTsconfigPathAliasWithoutCompilerOptions tests parsing without compilerOptions.
func TestParseTsconfigPathAliasWithoutCompilerOptions(t *testing.T) {
	buildOptions := &api.BuildOptions{
		TsconfigRaw:   `{"extends": "./base.json"}`,
		AbsWorkingDir: "/project",
	}

	pathAlias, err := parseTsconfigPathAlias(buildOptions)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(pathAlias) != 0 {
		t.Errorf("Expected empty path alias map without compilerOptions, got %d entries", len(pathAlias))
	}
}

// TestParseTsconfigPathAliasWithoutPaths tests parsing without paths in compilerOptions.
func TestParseTsconfigPathAliasWithoutPaths(t *testing.T) {
	buildOptions := &api.BuildOptions{
		TsconfigRaw: `{
            "compilerOptions": {
                "target": "ES2020",
                "module": "ESNext"
            }
        }`,
		AbsWorkingDir: "/project",
	}

	pathAlias, err := parseTsconfigPathAlias(buildOptions)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(pathAlias) != 0 {
		t.Errorf("Expected empty path alias map without paths, got %d entries", len(pathAlias))
	}
}
