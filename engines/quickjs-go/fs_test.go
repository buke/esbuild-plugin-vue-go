// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package qjscompiler

import (
	"os"
	"path/filepath"
	"testing"

	quickjsengine "github.com/buke/js-executor/engines/quickjs-go"
	"github.com/buke/quickjs-go"
)

// TestFileExistsFunc tests the fileExistsFunc function
func TestFileExistsFunc(t *testing.T) {
	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	// Create a temporary file for testing
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{"existing_file", tmpFile, true},
		{"nonexistent_file", "/nonexistent/file.txt", false},
		{"empty_path", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create function arguments
			args := []*quickjs.Value{ctx.String(test.filePath)}
			defer args[0].Free()

			// Call the function
			result := fileExistsFunc(ctx, nil, args)
			defer result.Free()

			// Check the result
			if result.Bool() != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result.Bool())
			}
		})
	}
}

// TestReadFileFunc tests the readFileFunc function
func TestReadFileFunc(t *testing.T) {
	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	// Create a temporary file with test content
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	testContent := "Hello, World!\nThis is a test file."
	err := os.WriteFile(tmpFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test successful file reading
	t.Run("existing_file", func(t *testing.T) {
		args := []*quickjs.Value{ctx.String(tmpFile)}
		defer args[0].Free()

		result := readFileFunc(ctx, nil, args)
		defer result.Free()

		if result.IsUndefined() {
			t.Error("Expected file content, got undefined")
		} else if result.String() != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, result.String())
		}
	})

	// Test error cases - these should throw exceptions
	errorTests := []struct {
		name string
		path string
	}{
		{"nonexistent_file", "/nonexistent/file.txt"},
		{"empty_path", ""},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			args := []*quickjs.Value{ctx.String(test.path)}
			defer args[0].Free()

			result := readFileFunc(ctx, nil, args)
			defer result.Free()

			// Check if an exception was thrown by checking if the result is an exception
			if !result.IsException() {
				t.Errorf("Expected exception for path '%s', but no exception was thrown", test.path)
			}
		})
	}
}

// TestRealpathFunc tests the realpathFunc function
func TestRealpathFunc(t *testing.T) {
	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test successful cases
	successTests := []struct {
		name      string
		inputPath string
	}{
		{"existing_file", tmpFile},
		{"existing_directory", tmpDir},
		{"relative_path", "."},
	}

	for _, test := range successTests {
		t.Run(test.name, func(t *testing.T) {
			args := []*quickjs.Value{ctx.String(test.inputPath)}
			defer args[0].Free()

			result := realpathFunc(ctx, nil, args)
			defer result.Free()

			if result.IsException() {
				t.Errorf("Expected successful result, got exception")
			} else if result.IsUndefined() {
				t.Error("Expected absolute path, got undefined")
			} else {
				resultPath := result.String()
				if !filepath.IsAbs(resultPath) {
					t.Errorf("Expected absolute path, got relative path: %s", resultPath)
				}
			}
		})
	}

	// Test error case - should throw exception
	t.Run("nonexistent_file", func(t *testing.T) {
		args := []*quickjs.Value{ctx.String("/nonexistent/file.txt")}
		defer args[0].Free()

		result := realpathFunc(ctx, nil, args)
		defer result.Free()

		// Check if an exception was thrown
		if !result.IsException() {
			t.Errorf("Expected exception for nonexistent file, but no exception was thrown")
		}
	})
}

// TestRealpathFuncWithRelativePath tests realpath with various relative paths
func TestRealpathFuncWithRelativePath(t *testing.T) {
	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	// Test with current directory
	args := []*quickjs.Value{ctx.String(".")}
	defer args[0].Free()

	result := realpathFunc(ctx, nil, args)
	defer result.Free()

	if result.IsException() {
		t.Error("Expected successful result, got exception")
	} else if result.IsUndefined() {
		t.Error("Expected absolute path for current directory, got undefined")
	}

	absolutePath := result.String()
	if !filepath.IsAbs(absolutePath) {
		t.Errorf("Expected absolute path, got: %s", absolutePath)
	}

	// Verify it's a valid directory path
	if stat, err := os.Stat(absolutePath); err != nil || !stat.IsDir() {
		t.Errorf("Expected valid directory path, got: %s (error: %v)", absolutePath, err)
	}
}

// TestLoadFsModule tests the loadFsModule function
func TestLoadFsModule(t *testing.T) {
	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	// Create a simple engine wrapper for testing
	engine := &quickjsengine.Engine{
		Runtime: runtime,
		Ctx:     ctx,
	}

	// Load the FS module
	err := loadFsModule(engine)
	if err != nil {
		t.Fatalf("Failed to load FS module: %v", err)
	}

	// Test that compilerFs object is available
	result := ctx.Eval("typeof compilerFs")
	defer result.Free()

	if result.String() != "object" {
		t.Errorf("Expected compilerFs to be an object, got: %s", result.String())
	}

	// Test that all expected functions are available
	expectedFunctions := []string{"fileExists", "readFile", "realpath"}
	for _, funcName := range expectedFunctions {
		funcResult := ctx.Eval("typeof compilerFs." + funcName)
		defer funcResult.Free()

		if funcResult.String() != "function" {
			t.Errorf("Expected compilerFs.%s to be a function, got: %s", funcName, funcResult.String())
		}
	}
}

// TestLoadFsModuleErrorHandling tests error handling in FS module functions
func TestLoadFsModuleErrorHandling(t *testing.T) {
	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	// Create a simple engine wrapper for testing
	engine := &quickjsengine.Engine{
		Runtime: runtime,
		Ctx:     ctx,
	}

	// Load the FS module
	err := loadFsModule(engine)
	if err != nil {
		t.Fatalf("Failed to load FS module: %v", err)
	}

	// Test fileExists with nonexistent file
	existsResult := ctx.Eval(`compilerFs.fileExists("/absolutely/nonexistent/path/file.txt")`)
	defer existsResult.Free()
	if existsResult.Bool() {
		t.Error("Expected nonexistent file to return false")
	}

	// Test readFile with nonexistent file - should throw exception
	readErrorTest := ctx.Eval(`
        try {
            compilerFs.readFile("/absolutely/nonexistent/path/file.txt");
            "no_error";
        } catch (err) {
            "error_thrown";
        }
    `)
	defer readErrorTest.Free()
	if readErrorTest.String() != "error_thrown" {
		t.Error("Expected readFile to throw error for nonexistent file")
	}

	// Test realpath with nonexistent file - should throw exception
	realpathErrorTest := ctx.Eval(`
        try {
            compilerFs.realpath("/absolutely/nonexistent/path/file.txt");
            "no_error";
        } catch (err) {
            "error_thrown";
        }
    `)
	defer realpathErrorTest.Free()
	if realpathErrorTest.String() != "error_thrown" {
		t.Error("Expected realpath to throw error for nonexistent file")
	}

	// Test with empty string paths
	emptyExistsResult := ctx.Eval(`compilerFs.fileExists("")`)
	defer emptyExistsResult.Free()
	if emptyExistsResult.Bool() {
		t.Error("Expected empty path to return false")
	}

	// Test readFile with empty path - should throw exception
	emptyReadErrorTest := ctx.Eval(`
        try {
            compilerFs.readFile("");
            "no_error";
        } catch (err) {
            "error_thrown";
        }
    `)
	defer emptyReadErrorTest.Free()
	if emptyReadErrorTest.String() != "error_thrown" {
		t.Error("Expected readFile to throw error for empty path")
	}
}

// TestRealpathFuncSymlinkResolution tests symbolic link resolution
func TestRealpathFuncSymlinkResolution(t *testing.T) {
	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	tmpDir := t.TempDir()

	// Create a target file
	targetFile := filepath.Join(tmpDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Create a symbolic link (skip test if symlinks not supported)
	linkFile := filepath.Join(tmpDir, "link.txt")
	err = os.Symlink(targetFile, linkFile)
	if err != nil {
		t.Skipf("Symbolic links not supported on this system: %v", err)
	}

	// Test that realpath resolves the symlink to the target
	args := []*quickjs.Value{ctx.String(linkFile)}
	defer args[0].Free()

	result := realpathFunc(ctx, nil, args)
	defer result.Free()

	if result.IsException() {
		t.Error("Expected successful result, got exception")
	} else if result.IsUndefined() {
		t.Error("Expected resolved path, got undefined")
	}

	resolvedPath := result.String()
	if !filepath.IsAbs(resolvedPath) {
		t.Errorf("Expected absolute path, got: %s", resolvedPath)
	}

	// The resolved path should point to the target file (absolute path)
	// Use filepath.EvalSymlinks to get the expected path to handle macOS /private prefix
	expectedPath, err := filepath.EvalSymlinks(targetFile)
	if err != nil {
		t.Fatalf("Failed to get real path of target: %v", err)
	}
	expectedPath, err = filepath.Abs(expectedPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path of target: %v", err)
	}

	if resolvedPath != expectedPath {
		t.Errorf("Expected resolved path '%s', got '%s'", expectedPath, resolvedPath)
	}
}
