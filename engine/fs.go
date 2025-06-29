// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"os"
	"path/filepath"

	quickjsengine "github.com/buke/js-executor/engines/quickjs-go"
	"github.com/buke/quickjs-go"
)

// fileExistsFunc checks if a file exists at the given path.
// Returns true if the file exists, otherwise false.
func fileExistsFunc(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
	file := args[0].String()
	if _, err := os.Stat(file); err != nil {
		return ctx.Bool(false)
	}
	return ctx.Bool(true)
}

// readFileFunc reads the content of a file and returns it as a string.
// If the file cannot be read, returns undefined.
func readFileFunc(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
	file := args[0].String()
	data, err := os.ReadFile(file)
	if err != nil {
		return ctx.Undefined()
	}
	return ctx.String(string(data))
}

// realpathFunc returns the absolute path of the given file.
// If the path cannot be resolved, returns undefined.
func realpathFunc(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
	file := args[0].String()
	realpath, err := filepath.Abs(file)
	if err != nil {
		return ctx.Undefined()
	}
	return ctx.String(realpath)
}

// loadFsModule injects a 'compilerFs' object into the JS context with file system helper functions.
// The object provides fileExists, readFile, and realpath methods for use in JS.
func loadFsModule(jse *quickjsengine.Engine) error {
	globalsObj := jse.Ctx.Globals()
	compilerFsObj := jse.Ctx.Object()
	compilerFsObj.Set("fileExists", jse.Ctx.Function(fileExistsFunc))
	compilerFsObj.Set("readFile", jse.Ctx.Function(readFileFunc))
	compilerFsObj.Set("realpath", jse.Ctx.Function(realpathFunc))
	globalsObj.Set("compilerFs", compilerFsObj)
	return nil
}
