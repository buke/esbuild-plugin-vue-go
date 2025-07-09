// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package qjscompiler

import (
	jsexecutor "github.com/buke/js-executor"
	quickjsengine "github.com/buke/js-executor/engines/quickjs-go"
)

// NewVueCompilerFactory creates a JsEngineFactory with Vue compiler, file system loaded.
// This factory is used to provide a ready-to-use JS engine for compiling Vue SFCs and handling related tasks.
// Additional QuickJS engine options can be passed via the variadic parameter.
func NewVueCompilerFactory(options ...quickjsengine.Option) jsexecutor.JsEngineFactory {
	// Inject file system helper functions into the JS context
	options = append(options, loadFsModule)
	// Load and evaluate the embedded Vue compiler JS bytecode
	options = append(options, loadCompilerModule)
	return quickjsengine.NewFactory(options...)
}
