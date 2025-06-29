// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package vueplugin

import (
	"fmt"
	"os"
	"path/filepath"

	jsexecutor "github.com/buke/js-executor"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/rs/xid"
)

// setupSassHandler registers handlers for Sass files (.scss and .sass).
func setupSassHandler(opts *options, build *api.PluginBuild) {
	// Register resolve handler for Sass files
	registerSassResolveHandler(opts, build)

	// Register load and compile handler for Sass files
	registerSassLoadHandler(opts, build)
}

// registerSassResolveHandler registers the path resolution handler for Sass files.
func registerSassResolveHandler(opts *options, build *api.PluginBuild) {
	build.OnResolve(api.OnResolveOptions{Filter: `\.s[ac]ss$`}, func(args api.OnResolveArgs) (api.OnResolveResult, error) {
		// Parse path aliases from tsconfig
		pathAlias, err := parseTsconfigPathAlias(build.InitialOptions)
		if err != nil {
			opts.logger.Error("Failed to parse tsconfig path aliases", "error", err)
			return api.OnResolveResult{}, err
		}

		args.Path = applyPathAlias(pathAlias, args.Path)

		// Handle relative paths
		path := args.Path
		if !filepath.IsAbs(args.Path) {
			path = filepath.Clean(filepath.Join(args.ResolveDir, args.Path))
		}

		return api.OnResolveResult{
			Path:      path,
			Namespace: "sass-loader",
		}, nil
	})
}

// registerSassLoadHandler registers the handler to load and compile Sass files.
func registerSassLoadHandler(opts *options, build *api.PluginBuild) {
	build.OnLoad(api.OnLoadOptions{Filter: `\.s[ac]ss$`, Namespace: "sass-loader"}, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
		// 1. Read the source content
		source, err := readSassSource(args, opts, build)
		if err != nil {
			opts.logger.Error("Failed to read Sass file", "error", err, "file", args.Path)
			return api.OnLoadResult{
				Errors: []api.Message{{
					Text: err.Error(),
					Location: &api.Location{
						File: args.Path,
					},
				}},
			}, err
		}

		// 2. Compile Sass to CSS
		css, err := compileSass(args.Path, source, opts.jsExecutor)
		if err != nil {
			opts.logger.Error("Failed to compile Sass", "error", err, "file", args.Path)
			return api.OnLoadResult{
				Errors: []api.Message{{
					Text: err.Error(),
					Location: &api.Location{
						File: args.Path,
					},
				}},
			}, err
		}

		return api.OnLoadResult{
			Contents: &css,
			Loader:   api.LoaderCSS,
		}, nil
	})
}

// readSassSource reads the Sass source file content.
// If any Sass load processor is present, it will be executed first.
// If no processor returns content, the file will be read from disk.
func readSassSource(args api.OnLoadArgs, opts *options, build *api.PluginBuild) (string, error) {
	// Execute Sass load processor chain if present
	if len(opts.onSassLoadProcessors) > 0 {
		for _, processor := range opts.onSassLoadProcessors {
			content, err := processor(args, build.InitialOptions)
			if err != nil {
				return "", fmt.Errorf("sass processor failed: %w", err)
			}
			if content != "" {
				return content, nil
			}
		}
	}

	// If no processor or all processors return empty, try to read the file from disk
	fbyte, err := os.ReadFile(args.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read sass file: %w", err)
	}

	return string(fbyte), nil
}

// compileSass compiles Sass to CSS using the Vue compiler via the JS executor.
func compileSass(filePath, source string, jsExecutor *jsexecutor.JsExecutor) (string, error) {
	location := filepath.Dir(filePath)

	jsResponse, err := jsExecutor.Execute(&jsexecutor.JsRequest{
		Id:      xid.New().String(),
		Service: "sfc.sass.renderSync",
		Args: []interface{}{map[string]interface{}{
			"data":         source,
			"sasslocation": location,
			"sourceMap":    false,
			"style":        "expanded",
		}},
	})

	if err != nil {
		return "", fmt.Errorf("sass compilation service failed: %w", err)
	}

	// Extract result
	result, ok := jsResponse.Result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response from sass compilation service")
	}

	// Get the compiled CSS code
	code, ok := result["css"].(string)
	if !ok {
		return "", fmt.Errorf("failed to extract CSS from compilation result")
	}

	return code, nil
}
