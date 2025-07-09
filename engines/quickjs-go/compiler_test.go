// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package qjscompiler

import (
	"testing"

	quickjsengine "github.com/buke/js-executor/engines/quickjs-go"
)

// TestNewVueCompilerFactory tests the NewVueCompilerFactory function
func TestNewVueCompilerFactory(t *testing.T) {
	// Test creating factory without additional options
	factory := NewVueCompilerFactory()
	if factory == nil {
		t.Fatal("Expected non-nil factory")
	}
}

// TestNewVueCompilerFactoryWithOptions tests creating factory with additional options
func TestNewVueCompilerFactoryWithOptions(t *testing.T) {
	// Create a custom option for testing
	customOptionCalled := false
	customOption := func(engine *quickjsengine.Engine) error {
		customOptionCalled = true
		return nil
	}

	// Create factory with custom option
	factory := NewVueCompilerFactory(customOption)
	if factory == nil {
		t.Fatal("Expected non-nil factory")
	}

	// Create an engine to verify the custom option was applied
	engine, err := factory()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Verify custom option was called
	if !customOptionCalled {
		t.Error("Expected custom option to be called")
	}
}

// TestVueCompilerFactoryEngineCreation tests engine creation from the factory
func TestVueCompilerFactoryEngineCreation(t *testing.T) {
	factory := NewVueCompilerFactory()

	// Create multiple engines to test factory functionality
	for i := 0; i < 3; i++ {
		engine, err := factory()
		if err != nil {
			t.Fatalf("Failed to create engine %d: %v", i, err)
		}

		// Verify it's a QuickJS engine
		qjsEngine, ok := engine.(*quickjsengine.Engine)
		if !ok {
			t.Errorf("Expected QuickJS engine, got %T", engine)
		}

		// Test basic engine functionality
		if qjsEngine.Runtime == nil {
			t.Error("Expected non-nil runtime")
		}
		if qjsEngine.Ctx == nil {
			t.Error("Expected non-nil context")
		}

		engine.Close()
	}
}
