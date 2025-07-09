// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package vueplugin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	jsexecutor "github.com/buke/js-executor"
	"github.com/cespare/xxhash"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/rs/xid"
)

// toPosixPath converts Windows-style paths to POSIX-style paths.
func toPosixPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// setupVueHandler registers all handlers for Vue files (.vue).
func setupVueHandler(opts *options, build *api.PluginBuild) {
	// Register main entry handler for .vue files
	registerMainEntryHandler(opts, build)

	// Register file resolve handler
	registerResolveHandler(opts, build)

	// Register handlers for Vue SFC parts
	registerScriptHandler(build)
	registerTemplateHandler(build)
	registerStyleHandler(build)
}

// registerMainEntryHandler processes .vue files and precompiles all SFC parts.
func registerMainEntryHandler(opts *options, build *api.PluginBuild) {
	isProd := true
	if v, exists := parseImportMetaEnv(build.InitialOptions.Define, "PROD"); exists {
		_v, ok := v.(bool)
		if ok {
			isProd = _v
		}
	}

	isSSR := false
	if v, exists := parseImportMetaEnv(build.InitialOptions.Define, "SSR"); exists {
		_v, ok := v.(bool)
		if ok {
			isSSR = _v
		}
	}

	build.OnLoad(api.OnLoadOptions{Filter: `\.vue$`}, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
		// 1. Read the source file content
		source, err := readVueSource(args, opts, build)
		if err != nil {
			opts.logger.Error("Failed to read Vue file", "error", err, "file", args.Path)
			return api.OnLoadResult{}, err
		}

		// 2. Generate component ID
		hashId := generateHashId(source)
		dataId := "data-v-" + hashId

		// 3. Compile SFC using the JS executor
		jsResponse, err := opts.jsExecutor.Execute(&jsexecutor.JsRequest{
			Id:      xid.New().String(),
			Service: "sfc.vue.compileSFC",
			Args: []interface{}{
				hashId,
				toPosixPath(args.Path),
				source,
				map[string]interface{}{
					"sourceMap":         build.InitialOptions.Sourcemap > 0,
					"isProd":            isProd,
					"isSSR":             isSSR,
					"preprocessOptions": opts.stylePreprocessorOptions,
					"compilerOptions":   opts.templateCompilerOptions,
				},
			},
		})
		if err != nil {
			opts.logger.Error("Failed to compile Vue SFC", "error", err, "file", args.Path)
			return api.OnLoadResult{
				Errors: []api.Message{{
					Text: fmt.Sprintf("Vue SFC compilation failed: %v", err),
					Location: &api.Location{
						File: args.Path,
					},
				}},
			}, err
		}

		compileResult, ok := jsResponse.Result.(map[string]interface{})
		if !ok {
			opts.logger.Error("Invalid Vue SFC compilation result", "result", jsResponse.Result, "file", args.Path)
			return api.OnLoadResult{
				Errors: []api.Message{{
					Text: fmt.Sprintf("Invalid Vue SFC compilation result: %v", jsResponse.Result),
					Location: &api.Location{
						File: args.Path,
					},
				}},
			}, err
		}

		// 4. Extract each part from the compilation result
		var script map[string]interface{}
		scriptResult, ok := compileResult["script"]
		if scriptResult != nil && ok {
			script = scriptResult.(map[string]interface{})
		}

		var template map[string]interface{}
		if templateResult, ok := compileResult["template"].(map[string]interface{}); ok && templateResult != nil {
			template = templateResult
		}

		var styles []map[string]interface{}
		if stylesResult, ok := compileResult["styles"].([]interface{}); ok && len(stylesResult) > 0 {
			styles = make([]map[string]interface{}, len(stylesResult))
			for i, s := range stylesResult {
				if sMap, ok := s.(map[string]interface{}); ok {
					styles[i] = sMap
				}
			}
		}

		// 5. Generate entry content - use compilation result directly, no need for descriptors
		contents, err := generateEntryContents(args.Path, dataId, isSSR, script, template, styles)
		if err != nil {
			opts.logger.Error("Failed to generate Vue entry contents", "error", err, "file", args.Path)
			return api.OnLoadResult{
				Errors: []api.Message{{
					Text: fmt.Sprintf("Failed to generate Vue entry contents: %v", err),
					Location: &api.Location{
						File: args.Path,
					},
				}},
			}, err
		}

		// 6. Prepare plugin data
		pluginData := map[string]interface{}{
			"id":     dataId,
			"script": script,
		}

		if template != nil {
			pluginData["template"] = template
		}

		if styles != nil {
			pluginData["styles"] = styles
		}

		// 7. Extract script warnings
		buildWarnings := make([]api.Message, 0)
		if warnings, ok := script["warnings"].([]interface{}); ok {
			for _, w := range warnings {
				if warn, isString := w.(string); isString {
					buildWarnings = append(buildWarnings, api.Message{Text: warn})
				}
			}
		}

		return api.OnLoadResult{
			Contents:   &contents,
			ResolveDir: filepath.Dir(args.Path),
			PluginData: pluginData,
			Warnings:   buildWarnings,
		}, nil
	})
}

// readVueSource reads and preprocesses the Vue source file.
// It also executes the Vue load processor chain if present.
func readVueSource(args api.OnLoadArgs, opts *options, build *api.PluginBuild) (string, error) {
	fbyte, err := os.ReadFile(args.Path)
	if err != nil {
		return "", err
	}

	source := string(fbyte)
	source = strings.ReplaceAll(source, "\r\n", "\n")

	// Execute Vue load processor chain if any
	if len(opts.onVueLoadProcessors) > 0 {
		for _, processor := range opts.onVueLoadProcessors {
			var processorErr error
			source, processorErr = processor(source, args, build.InitialOptions)
			if processorErr != nil {
				return "", processorErr
			}
		}
	}

	return source, nil
}

// generateHashId generates a hash ID for the given source string.
func generateHashId(source string) string {
	return strconv.FormatUint(xxhash.Sum64String(source), 16)
}

// generateEntryContents generates the entry JS code for a Vue SFC.
// It uses a Go template to generate the code based on the SFC parts.
func generateEntryContents(filePath, dataId string, isSSR bool,
	sfcScript map[string]interface{}, sfcTemplate map[string]interface{},
	sfcStyles []map[string]interface{}) (string, error) {

	// Convert filePath to relative POSIX path for import statements
	relPath, err := filepath.Rel(".", filePath)
	if err != nil {
		relPath = filePath
	}
	relPath = toPosixPath(relPath)

	tpl := template.Must(template.New(filePath).Funcs(template.FuncMap{
		"SSR": func() bool {
			return isSSR
		},
		"dataId": func() string {
			return dataId
		},
		"hasScript": func() bool {
			return sfcScript != nil && sfcScript["content"] != nil
		},
		"hasTemplate": func() bool {
			return sfcTemplate != nil && sfcTemplate["code"] != nil
		},
		"someScoped": func() bool {
			for _, style := range sfcStyles {
				if scoped := style["scoped"].(bool); scoped {
					return true
				}
			}
			return false
		},
	}).Parse(`
{{ if hasScript }}
import script from '{{ .relPath }}?type=script'
{{ else }}
const script = {};
{{ end }}

{{ range $index, $_ := .styles }}
import '{{ $.relPath }}?type=style&index={{ $index }}'
{{ end }}

{{ if hasTemplate }}
import { {{ if SSR }}ssrRender{{ else }}render{{ end }} } from '{{ .relPath }}?type=template'
script.{{ if SSR }}ssrRender{{ else }}render{{ end }} = {{ if SSR }}ssrRender{{ else }}render{{ end }};
{{ end }}

script.__file = {{ .relPath | printf "%q" }};
{{ if someScoped }}
script.__scopeId = {{ printf "%q" (dataId) }};{{ end }}
{{ if SSR }}
script.__ssrInlineRender = true;
{{ end }}

{{ if hasScript }}
export * from '{{ .relPath }}?type=script'
{{ end }}
export default script;
`))

	data := map[string]interface{}{
		"relPath": relPath,
		"styles":  sfcStyles,
	}

	contentsBuf := new(bytes.Buffer)
	if err := tpl.Execute(contentsBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute Vue entry template: %w", err)
	}

	return contentsBuf.String(), nil
}

// registerResolveHandler registers the file resolution handler for .vue files.
// Handles path aliases, relative paths, and parses URL parameters.
func registerResolveHandler(opts *options, build *api.PluginBuild) {
	build.OnResolve(api.OnResolveOptions{Filter: `\.vue(\?.*)?$`}, func(args api.OnResolveArgs) (api.OnResolveResult, error) {
		// Handle path aliases
		pathAlias, err := parseTsconfigPathAlias(build.InitialOptions)
		if err != nil {
			return api.OnResolveResult{}, err
		}
		args.Path = applyPathAlias(pathAlias, args.Path)

		// Handle relative paths
		if !filepath.IsAbs(args.Path) {
			args.Path = filepath.Clean(filepath.Join(args.ResolveDir, args.Path))
		}

		// Execute Vue resolve processor chain if any
		if len(opts.onVueResolveProcessors) > 0 {
			for _, processor := range opts.onVueResolveProcessors {
				result, err := processor(&args, build.InitialOptions)
				if err != nil {
					return api.OnResolveResult{}, err
				}
				if result != nil {
					return *result, nil
				}
			}
		}

		// Parse URL parameters
		pathUrl, _ := url.Parse(args.Path)
		params := pathUrl.Query()
		namespace := "file"
		if t := params.Get("type"); t != "" {
			namespace = "sfc-" + t
		}

		// Setup plugin data
		if args.PluginData == nil {
			args.PluginData = make(map[string]interface{})
		}

		if indexStr := params.Get("index"); indexStr != "" {
			if index, err := strconv.Atoi(indexStr); err == nil {
				args.PluginData.(map[string]interface{})["index"] = index
			}
		}

		return api.OnResolveResult{
			Path:       args.Path,
			Namespace:  namespace,
			PluginData: args.PluginData,
		}, nil
	})
}

// registerScriptHandler registers the script handler for .vue files.
// Loads the script part and attaches sourcemap if present.
func registerScriptHandler(build *api.PluginBuild) {
	build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: "sfc-script"}, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
		pluginData := args.PluginData.(map[string]interface{})
		script := pluginData["script"].(map[string]interface{})

		content := script["content"].(string)

		// Add sourcemap if present
		if build.InitialOptions.Sourcemap > 0 && script["map"] != nil {
			sourceMapJSON, err := json.Marshal(script["map"])
			if err != nil {
				return api.OnLoadResult{}, err
			}
			sourceMapBase64 := base64.StdEncoding.EncodeToString(sourceMapJSON)
			content += "\n\n//@ sourceMappingURL=data:application/json;charset=utf-8;base64," + sourceMapBase64
		}

		// Determine loader type
		loader := api.LoaderJS
		if script["lang"] != nil && script["lang"].(string) == "ts" {
			loader = api.LoaderTS
		}

		return api.OnLoadResult{
			Contents:   &content,
			Loader:     loader,
			ResolveDir: filepath.Dir(args.Path),
			PluginData: pluginData,
		}, nil
	})
}

// registerTemplateHandler registers the template handler for .vue files.
// Loads the precompiled template part and attaches warnings if present.
func registerTemplateHandler(build *api.PluginBuild) {
	build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: "sfc-template"}, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
		pluginData := args.PluginData.(map[string]interface{})

		// Use precompiled template result
		templateResult := pluginData["template"].(map[string]interface{})

		code := templateResult["code"].(string)

		// Extract warning information
		var mappedTips []api.Message
		if tips, ok := templateResult["tips"].([]interface{}); ok {
			mappedTips = make([]api.Message, len(tips))
			for i, tip := range tips {
				mappedTips[i] = api.Message{
					Text: tip.(string),
				}
			}
		}

		return api.OnLoadResult{
			Contents:   &code,
			Warnings:   mappedTips,
			Loader:     api.LoaderTS,
			ResolveDir: filepath.Dir(args.Path),
			PluginData: pluginData,
		}, nil
	})
}

// registerStyleHandler registers the style handler for .vue files.
// Loads the precompiled style part based on the index from URL parameters.
func registerStyleHandler(build *api.PluginBuild) {
	build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: "sfc-style"}, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
		pluginData := args.PluginData.(map[string]interface{})

		// Get index from URL query parameters
		parsedURL, _ := url.Parse(args.Path)

		params := parsedURL.Query()
		indexStr := params.Get("index")
		index, _ := strconv.Atoi(indexStr)

		// Use precompiled style result
		styles := pluginData["styles"].([]map[string]interface{})

		styleResult := styles[index]
		code := styleResult["code"].(string)

		return api.OnLoadResult{
			Contents:   &code,
			Loader:     api.LoaderCSS,
			ResolveDir: filepath.Dir(args.Path),
			PluginData: pluginData,
		}, nil
	})
}
