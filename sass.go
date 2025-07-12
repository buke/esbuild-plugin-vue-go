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
// It sets up the complete processing pipeline including path resolution and compilation.
func setupSassHandler(opts *Options, build *api.PluginBuild) {
	// Register resolve handler for Sass files
	registerSassResolveHandler(opts, build)

	// Register load and compile handler for Sass files
	registerSassLoadHandler(opts, build)
}

// registerSassResolveHandler registers the path resolution handler for Sass files.
// It handles TypeScript path aliases and converts relative paths to absolute paths.
// Resolved Sass files are assigned to the "sass-loader" namespace for further processing.
func registerSassResolveHandler(opts *Options, build *api.PluginBuild) {
	build.OnResolve(api.OnResolveOptions{Filter: `\.s[ac]ss$`}, func(args api.OnResolveArgs) (api.OnResolveResult, error) {
		// Parse TypeScript path aliases from tsconfig to support project-wide path mapping
		pathAlias, err := parseTsconfigPathAlias(build.InitialOptions)
		if err != nil {
			opts.logger.Error("Failed to parse tsconfig path aliases", "error", err)
			return api.OnResolveResult{}, err
		}

		// Apply path aliases to support imports like @/styles/main.scss
		args.Path = applyPathAlias(pathAlias, args.Path)

		// Convert relative paths to absolute paths for consistent file resolution
		path := args.Path
		if !filepath.IsAbs(args.Path) {
			path = filepath.Clean(filepath.Join(args.ResolveDir, args.Path))
		}

		return api.OnResolveResult{
			Path:      path,
			Namespace: "sass-loader", // Assign to sass-loader namespace for compilation
		}, nil
	})
}

// registerSassLoadHandler registers the handler to load and compile Sass files.
// It reads the source content, processes it through any registered processors,
// and compiles it to CSS using the Vue compiler's Sass service.
func registerSassLoadHandler(opts *Options, build *api.PluginBuild) {
	build.OnLoad(api.OnLoadOptions{Filter: `\.s[ac]ss$`, Namespace: "sass-loader"}, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
		// Step 1: Read the Sass source content (with optional preprocessing)
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

		// Step 2: Compile Sass to CSS using the Vue compiler's integrated Sass service
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

		// Step 3: Return compiled CSS with appropriate loader
		return api.OnLoadResult{
			Contents: &css,
			Loader:   api.LoaderCSS, // Use CSS loader for the compiled output
		}, nil
	})
}

// readSassSource reads the Sass source file content with optional preprocessing.
// If any Sass load processor is registered, it will be executed first to allow
// custom content transformation or dynamic content generation.
// If no processor returns content, the file will be read from disk.
func readSassSource(args api.OnLoadArgs, opts *Options, build *api.PluginBuild) (string, error) {
	// Execute Sass load processor chain if present
	// This allows for custom preprocessing, content injection, or dynamic imports
	if len(opts.onSassLoadProcessors) > 0 {
		for _, processor := range opts.onSassLoadProcessors {
			content, err := processor(args, build.InitialOptions)
			if err != nil {
				return "", fmt.Errorf("sass processor failed: %w", err)
			}
			// If processor returns non-empty content, use it instead of file content
			if content != "" {
				return content, nil
			}
		}
	}

	// Fallback: read the file from disk if no processor provides content
	fbyte, err := os.ReadFile(args.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read sass file: %w", err)
	}

	return string(fbyte), nil
}

// compileSass compiles Sass to CSS using the Vue compiler via the JS executor.
// It uses the integrated Sass compiler service that supports both .scss and .sass syntax.
// The compilation includes dependency resolution and supports Sass features like imports,
// variables, mixins, and functions.
func compileSass(filePath, source string, jsExecutor *jsexecutor.JsExecutor) (string, error) {
	// Extract directory path for Sass import resolution
	location := filepath.Dir(filePath)

	// Execute Sass compilation via the Vue compiler's Sass service
	jsResponse, err := jsExecutor.Execute(&jsexecutor.JsRequest{
		Id:      xid.New().String(),
		Service: "sfc.sass.renderSync", // Vue compiler's integrated Sass service
		Args: []interface{}{map[string]interface{}{
			"data":         source,     // Sass source code to compile
			"sasslocation": location,   // Base directory for resolving @import statements
			"sourceMap":    false,      // Disable source maps for production builds
			"style":        "expanded", // Output style: expanded, compressed, etc.
		}},
	})

	if err != nil {
		return "", fmt.Errorf("sass compilation service failed: %w", err)
	}

	// Extract and validate compilation result
	result, ok := jsResponse.Result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response from sass compilation service")
	}

	// Extract the compiled CSS code from the result
	code, ok := result["css"].(string)
	if !ok {
		return "", fmt.Errorf("failed to extract CSS from compilation result")
	}

	return code, nil
}
