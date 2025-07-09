// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package qjscompiler

import (
	_ "embed"
	"sync"

	quickjsengine "github.com/buke/js-executor/engines/quickjs-go"
	quickjs "github.com/buke/quickjs-go"
)

// compilerScript embeds the Vue compiler JavaScript code.
//
//go:embed compilerjs/dist/index.js
var compilerScript string

var (
	once                   sync.Once
	compilerScriptBytecode []byte
)

// getCompilerScriptBytecode compiles the embedded JS compiler script and caches its bytecode.
// It uses sync.Once to ensure compilation happens only once per process.
func getCompilerScriptBytecode(jse *quickjsengine.Engine) []byte {
	once.Do(func() {
		b, err := jse.Ctx.Compile(compilerScript, quickjs.EvalFileName("compilerjs/dist/index.js"))
		if err != nil {
			panic(err)
		}
		compilerScriptBytecode = b
	})
	return compilerScriptBytecode
}

// loadCompilerModule loads and evaluates the compiled JS compiler bytecode in the QuickJS context.
func loadCompilerModule(jse *quickjsengine.Engine) error {
	// Get the compiled bytecode for the compiler script
	bytecode := getCompilerScriptBytecode(jse)

	// Evaluate the bytecode to load the compiler module into the JS context
	ret := jse.Ctx.EvalBytecode(bytecode)
	defer ret.Free()

	if ret.IsException() {
		return jse.Ctx.Exception()
	}

	return nil
}
