// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package qjscompiler

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
func readFileFunc(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
	file := args[0].String()
	data, err := os.ReadFile(file)
	if err != nil {
		return ctx.ThrowError(err)
	}
	return ctx.String(string(data))
}

// realpathFunc returns the real absolute path of the given file, resolving symbolic links.
func realpathFunc(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
	file := args[0].String()

	// EvalSymlinks resolves symlinks and validates existence
	resolved, err := filepath.EvalSymlinks(file)
	if err != nil {
		return ctx.ThrowError(err)
	}

	// Convert to absolute path (this should rarely fail since path exists)
	realpath, _ := filepath.Abs(resolved)

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
