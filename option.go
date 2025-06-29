// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package vueplugin

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	jsexecutor "github.com/buke/js-executor"
	"github.com/evanw/esbuild/pkg/api"
	"golang.org/x/net/html"
)

// OnStartProcessor is a function type for processing logic before the build starts.
// Receives the esbuild BuildOptions as input.
type OnStartProcessor func(buildOptions *api.BuildOptions) error

// OnVueResolveProcessor is a function type for custom Vue file resolution logic.
// Receives the OnResolveArgs and BuildOptions, returns an optional OnResolveResult and error.
type OnVueResolveProcessor func(args *api.OnResolveArgs, buildOptions *api.BuildOptions) (*api.OnResolveResult, error)

// OnVueLoadProcessor is a function type for custom Vue file loading logic.
// Receives the file content, OnLoadArgs, and BuildOptions, returns the (possibly transformed) content and error.
type OnVueLoadProcessor func(content string, args api.OnLoadArgs, buildOptions *api.BuildOptions) (string, error)

// OnSassLoadProcessor is a function type for custom Sass file loading logic.
// Receives the OnLoadArgs and BuildOptions, returns the file content and error.
type OnSassLoadProcessor func(args api.OnLoadArgs, buildOptions *api.BuildOptions) (content string, err error)

// OnEndProcessor is a function type for processing logic after the build ends.
// Receives the BuildResult and BuildOptions as input.
type OnEndProcessor func(result *api.BuildResult, buildOptions *api.BuildOptions) error

// OnDisposeProcessor is a function type for cleanup logic after the build is disposed.
// Receives the BuildOptions as input.
type OnDisposeProcessor func(buildOptions *api.BuildOptions) error

// IndexHtmlProcessor is a function type for processing HTML files after build.
// Receives the HTML document node, BuildResult, plugin options, and PluginBuild context.
type IndexHtmlProcessor func(doc *html.Node, result *api.BuildResult, opts *options, build *api.PluginBuild) error

// IndexHtmlOptions holds options for HTML processing.
type IndexHtmlOptions struct {
	SourceFile          string   // Source HTML file
	OutFile             string   // Output HTML file
	RemoveTagXPaths     []string // XPath expressions for removing specific HTML nodes
	IndexHtmlProcessors []IndexHtmlProcessor
}

// options holds all plugin configuration and processor chains.
type options struct {
	name                     string
	templateCompilerOptions  map[string]any
	stylePreprocessorOptions map[string]any
	indexHtmlOptions         IndexHtmlOptions

	// Processor chains for plugin extension points
	onStartProcessors      []OnStartProcessor
	onVueResolveProcessors []OnVueResolveProcessor
	onVueLoadProcessors    []OnVueLoadProcessor
	onSassLoadProcessors   []OnSassLoadProcessor
	onEndProcessors        []OnEndProcessor
	onDisposeProcessors    []OnDisposeProcessor

	jsExecutor *jsexecutor.JsExecutor
	logger     *slog.Logger
}

// OptionFunc is a function type for configuring plugin options.
type OptionFunc func(*options)

// newOptions creates a new options struct with default values.
func newOptions() *options {
	return &options{
		name:                     "vue-plugin",
		templateCompilerOptions:  make(map[string]any),
		stylePreprocessorOptions: make(map[string]any),
		logger:                   slog.Default(),
	}
}

// WithName sets the plugin name.
func WithName(name string) func(*options) {
	return func(opts *options) {
		opts.name = name
	}
}

// WithTemplateCompilerOptions sets the template compiler options.
func WithTemplateCompilerOptions(templateCompilerOptions map[string]any) func(*options) {
	return func(opts *options) {
		opts.templateCompilerOptions = templateCompilerOptions
	}
}

// WithStylePreprocessorOptions sets the style preprocessor options.
func WithStylePreprocessorOptions(stylePreprocessorOptions map[string]any) func(*options) {
	return func(opts *options) {
		opts.stylePreprocessorOptions = stylePreprocessorOptions
	}
}

// WithIndexHtmlOptions sets the HTML processing options.
func WithIndexHtmlOptions(indexHtmlOptions IndexHtmlOptions) func(*options) {
	return func(opts *options) {
		opts.indexHtmlOptions = indexHtmlOptions
	}
}

// WithOnStartProcessor adds an OnStartProcessor to the processor chain.
func WithOnStartProcessor(processor OnStartProcessor) func(*options) {
	return func(opts *options) {
		opts.onStartProcessors = append(opts.onStartProcessors, processor)
	}
}

// WithOnVueResolveProcessor adds an OnVueResolveProcessor to the processor chain.
func WithOnVueResolveProcessor(processor OnVueResolveProcessor) func(*options) {
	return func(opts *options) {
		opts.onVueResolveProcessors = append(opts.onVueResolveProcessors, processor)
	}
}

// WithOnVueLoadProcessor adds an OnVueLoadProcessor to the processor chain.
func WithOnVueLoadProcessor(processor OnVueLoadProcessor) func(*options) {
	return func(opts *options) {
		opts.onVueLoadProcessors = append(opts.onVueLoadProcessors, processor)
	}
}

// WithOnSassLoadProcessor adds an OnSassLoadProcessor to the processor chain.
func WithOnSassLoadProcessor(processor OnSassLoadProcessor) func(*options) {
	return func(opts *options) {
		opts.onSassLoadProcessors = append(opts.onSassLoadProcessors, processor)
	}
}

// WithOnEndProcessor adds an OnEndProcessor to the processor chain.
func WithOnEndProcessor(processor OnEndProcessor) func(*options) {
	return func(opts *options) {
		opts.onEndProcessors = append(opts.onEndProcessors, processor)
	}
}

// WithOnDisposeProcessor adds an OnDisposeProcessor to the processor chain.
func WithOnDisposeProcessor(processor OnDisposeProcessor) func(*options) {
	return func(opts *options) {
		opts.onDisposeProcessors = append(opts.onDisposeProcessors, processor)
	}
}

// WithJsExecutor sets the JS executor for the plugin.
func WithJsExecutor(jsExecutor *jsexecutor.JsExecutor) func(*options) {
	return func(opts *options) {
		opts.jsExecutor = jsExecutor
	}
}

// WithLogger sets the logger for the plugin.
func WithLogger(logger *slog.Logger) func(*options) {
	return func(opts *options) {
		opts.logger = logger
	}
}

// parseImportMetaEnv parses import.meta.env values from esbuild Define map.
func parseImportMetaEnv(defineMap map[string]string, key string) (any, bool) {
	if v, ok := defineMap[fmt.Sprintf("import.meta.env.%s", key)]; ok {
		var value interface{}
		json.Unmarshal([]byte(v), &value)
		return value, true
	}

	if v, ok := defineMap["import.meta.env"]; ok {
		var value interface{}
		json.Unmarshal([]byte(v), &value)
		if envMap, ok := value.(map[string]interface{}); ok {
			if v, ok := envMap[key]; ok {
				return v, true
			}
		}
	}

	return nil, false
}

// normalizeEsbuildOptions sets default values and ensures esbuild options are valid.
func normalizeEsbuildOptions(initialOptions *api.BuildOptions) {

	if initialOptions.Define == nil {
		initialOptions.Define = make(map[string]string)
	}

	// Set default import.meta.env
	if _, ok := initialOptions.Define["import.meta.env"]; !ok {
		initialOptions.Define["import.meta.env"] = "{}"
	}

	if _, exits := parseImportMetaEnv(initialOptions.Define, "MODE"); !exits {
		initialOptions.Define["import.meta.env.MODE"] = "'production'"
	}
	if _, exits := parseImportMetaEnv(initialOptions.Define, "PROD"); !exits {
		initialOptions.Define["import.meta.env.PROD"] = "true"
	}
	if _, exits := parseImportMetaEnv(initialOptions.Define, "DEV"); !exits {
		initialOptions.Define["import.meta.env.DEV"] = "false"
	}
	if _, exits := parseImportMetaEnv(initialOptions.Define, "SSR"); !exits {
		initialOptions.Define["import.meta.env.SSR"] = "false"
	}
	if _, exits := parseImportMetaEnv(initialOptions.Define, "BASE_URL"); !exits {
		initialOptions.Define["import.meta.env.BASE_URL"] = "'/'"
	}

	if _, ok := initialOptions.Define["__VUE_OPTIONS_API__"]; !ok {
		initialOptions.Define["__VUE_OPTIONS_API__"] = fmt.Sprintf(`%t`, true)
	}
	if _, ok := initialOptions.Define["__VUE_PROD_DEVTOOLS__"]; !ok {
		initialOptions.Define["__VUE_PROD_DEVTOOLS__"] = fmt.Sprintf(`%t`, false)
	}
	if _, ok := initialOptions.Define["__VUE_PROD_HYDRATION_MISMATCH_DETAILS__"]; !ok {
		initialOptions.Define["__VUE_PROD_HYDRATION_MISMATCH_DETAILS__"] = fmt.Sprintf(`%t`, false)
	}

	initialOptions.Metafile = true

}

// SimpleCopy returns an OnEndProcessor that copies files from fileMap after build.
// Each key-value pair in fileMap is srcFile -> outFile.
func SimpleCopy(fileMap map[string]string) OnEndProcessor {
	return func(result *api.BuildResult, initialOptions *api.BuildOptions) error {
		// Copy the output files to the desired location
		for srcFile, outFile := range fileMap {
			// Open source file
			src, err := os.Open(srcFile)
			if err != nil {
				return fmt.Errorf("failed to open source file %s: %w", srcFile, err)
			}
			defer src.Close()

			// Ensure the output directory exists
			if err := os.MkdirAll(filepath.Dir(outFile), 0755); err != nil {
				return fmt.Errorf("failed to create output dir for %s: %w", outFile, err)
			}

			// Create the output file
			dst, err := os.Create(outFile)
			if err != nil {
				return fmt.Errorf("failed to create output file %s: %w", outFile, err)
			}
			defer dst.Close()

			// Copy file content
			if _, err := io.Copy(dst, src); err != nil {
				return fmt.Errorf("failed to copy from %s to %s: %w", srcFile, outFile, err)
			}
		}
		return nil
	}
}
