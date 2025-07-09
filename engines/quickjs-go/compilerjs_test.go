// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package qjscompiler

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	quickjsengine "github.com/buke/js-executor/engines/quickjs-go"
	"github.com/buke/quickjs-go"
)

// TestCompilerScriptEmbedded tests that the embedded compiler script is not empty
func TestCompilerScriptEmbedded(t *testing.T) {
	if compilerScript == "" {
		t.Error("Expected non-empty embedded compiler script")
	}

	// Basic sanity check that it looks like JavaScript
	if len(compilerScript) < 100 {
		t.Error("Embedded compiler script seems too short")
	}
}

// TestGetCompilerScriptBytecode tests the getCompilerScriptBytecode function
func TestGetCompilerScriptBytecode(t *testing.T) {
	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	engine := &quickjsengine.Engine{
		Runtime: runtime,
		Ctx:     ctx,
	}

	// Reset the once variable for testing
	resetOnceForTesting()

	// First call should compile and cache the bytecode
	bytecode1 := getCompilerScriptBytecode(engine)
	if len(bytecode1) == 0 {
		t.Error("Expected non-empty bytecode")
	}

	// Second call should return the same cached bytecode
	bytecode2 := getCompilerScriptBytecode(engine)
	if !reflect.DeepEqual(bytecode1, bytecode2) {
		t.Error("Expected the same bytecode from cache")
	}

	// Verify the global variable is set
	if len(compilerScriptBytecode) == 0 {
		t.Error("Expected global bytecode cache to be populated")
	}
}

// TestGetCompilerScriptBytecodeConcurrency tests concurrent access to getCompilerScriptBytecode
func TestGetCompilerScriptBytecodeConcurrency(t *testing.T) {
	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	engine := &quickjsengine.Engine{
		Runtime: runtime,
		Ctx:     ctx,
	}

	// Reset the once variable for testing
	resetOnceForTesting()

	const numGoroutines = 10
	results := make([][]byte, numGoroutines)
	var wg sync.WaitGroup

	// Launch multiple goroutines to test concurrent access
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = getCompilerScriptBytecode(engine)
		}(i)
	}

	wg.Wait()

	// All results should be identical (same cached bytecode)
	firstResult := results[0]
	if len(firstResult) == 0 {
		t.Error("Expected non-empty bytecode")
	}

	for i := 1; i < numGoroutines; i++ {
		if !reflect.DeepEqual(firstResult, results[i]) {
			t.Errorf("Bytecode mismatch at index %d", i)
		}
	}
}

// TestLoadCompilerModule tests the loadCompilerModule function
func TestLoadCompilerModule(t *testing.T) {
	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	engine := &quickjsengine.Engine{
		Runtime: runtime,
		Ctx:     ctx,
	}

	// Reset the once variable for testing
	resetOnceForTesting()

	// Load the compiler module
	err := loadCompilerModule(engine)
	if err != nil {
		t.Fatalf("Failed to load compiler module: %v", err)
	}

	// Test that the sfc object is available in the global scope
	result := ctx.Eval("typeof sfc")
	defer result.Free()

	if result.String() != "object" {
		t.Errorf("Expected sfc to be an object, got: %s", result.String())
	}

	// Test that sfc.vue is available
	vueResult := ctx.Eval("typeof sfc.vue")
	defer vueResult.Free()

	if vueResult.String() != "object" {
		t.Errorf("Expected sfc.vue to be an object, got: %s", vueResult.String())
	}

	// Test that sfc.vue.compileSFC function is available
	compileSFCResult := ctx.Eval("typeof sfc.vue.compileSFC")
	defer compileSFCResult.Free()

	if compileSFCResult.String() != "function" {
		t.Errorf("Expected sfc.vue.compileSFC to be a function, got: %s", compileSFCResult.String())
	}
}

// TestLoadCompilerModuleMultipleTimes tests loading the compiler module multiple times
func TestLoadCompilerModuleMultipleTimes(t *testing.T) {
	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	engine := &quickjsengine.Engine{
		Runtime: runtime,
		Ctx:     ctx,
	}

	// Reset the once variable for testing
	resetOnceForTesting()

	// Load the compiler module multiple times
	for i := 0; i < 3; i++ {
		err := loadCompilerModule(engine)
		if err != nil {
			t.Fatalf("Failed to load compiler module on iteration %d: %v", i, err)
		}

		// Verify that sfc.vue.compileSFC is still available
		result := ctx.Eval("typeof sfc.vue.compileSFC")
		defer result.Free()

		if result.String() != "function" {
			t.Errorf("Expected sfc.vue.compileSFC to be a function on iteration %d, got: %s", i, result.String())
		}
	}
}

// TestLoadCompilerModuleWithDifferentEngines tests loading compiler module with different engines
func TestLoadCompilerModuleWithDifferentEngines(t *testing.T) {
	// Reset the once variable for testing
	resetOnceForTesting()

	// Create first engine
	runtime1 := quickjs.NewRuntime()
	defer runtime1.Close()
	ctx1 := runtime1.NewContext()
	defer ctx1.Close()
	engine1 := &quickjsengine.Engine{Runtime: runtime1, Ctx: ctx1}

	// Create second engine
	runtime2 := quickjs.NewRuntime()
	defer runtime2.Close()
	ctx2 := runtime2.NewContext()
	defer ctx2.Close()
	engine2 := &quickjsengine.Engine{Runtime: runtime2, Ctx: ctx2}

	// Load compiler module in both engines
	err1 := loadCompilerModule(engine1)
	if err1 != nil {
		t.Fatalf("Failed to load compiler module in engine1: %v", err1)
	}

	err2 := loadCompilerModule(engine2)
	if err2 != nil {
		t.Fatalf("Failed to load compiler module in engine2: %v", err2)
	}

	// Both engines should have the sfc object available
	result1 := ctx1.Eval("typeof sfc.vue.compileSFC")
	defer result1.Free()
	result2 := ctx2.Eval("typeof sfc.vue.compileSFC")
	defer result2.Free()

	if result1.String() != "function" {
		t.Error("Expected sfc.vue.compileSFC to be available in engine1")
	}

	if result2.String() != "function" {
		t.Error("Expected sfc.vue.compileSFC to be available in engine2")
	}
}

// TestBytecodeCompilationError tests error handling during bytecode compilation
func TestBytecodeCompilationError(t *testing.T) {
	// Save original compiler script
	originalScript := compilerScript

	// Temporarily replace with invalid JavaScript to trigger compilation error
	compilerScript = "invalid javascript syntax {"

	// Reset the once variable for testing
	resetOnceForTesting()

	defer func() {
		// Restore original script and reset once for cleanup
		compilerScript = originalScript
		resetOnceForTesting()

		// Recover from panic
		if r := recover(); r == nil {
			t.Error("Expected panic due to compilation error")
		}
	}()

	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	engine := &quickjsengine.Engine{
		Runtime: runtime,
		Ctx:     ctx,
	}

	// This should panic due to invalid JavaScript
	getCompilerScriptBytecode(engine)
}

// resetOnceForTesting resets the sync.Once variable for testing purposes
// This allows us to test the compilation process multiple times
func resetOnceForTesting() {
	once = sync.Once{}
	compilerScriptBytecode = nil
}

// TestLoadCompilerModuleException tests exception handling in loadCompilerModule
func TestLoadCompilerModuleException(t *testing.T) {
	runtime := quickjs.NewRuntime()
	defer runtime.Close()

	ctx := runtime.NewContext()
	defer ctx.Close()

	engine := &quickjsengine.Engine{
		Runtime: runtime,
		Ctx:     ctx,
	}

	// Save original compiler script
	originalScript := compilerScript

	// Temporarily replace with JavaScript that will cause runtime exception
	// This script compiles fine but throws an exception when executed
	compilerScript = `
        // This will compile successfully but throw a runtime exception
        throw new Error("Intentional runtime error for testing");
    `

	// Reset the once variable to force recompilation
	resetOnceForTesting()

	defer func() {
		// Restore original script and reset once for cleanup
		compilerScript = originalScript
		resetOnceForTesting()
	}()

	// This should return an error due to the runtime exception
	err := loadCompilerModule(engine)
	if err == nil {
		t.Error("Expected error due to runtime exception, got nil")
	}

	// Check that the error message contains our intentional error
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}

	// Optional: check for specific error content
	if !strings.Contains(err.Error(), "Error") {
		t.Logf("Error message: %s", err.Error())
	}
}
