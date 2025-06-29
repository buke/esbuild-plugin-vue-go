// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package vueplugin

import (
	"github.com/evanw/esbuild/pkg/api"
)

// NewPlugin creates a new esbuild plugin for Vue SFC support.
// Accepts a list of OptionFunc to customize plugin behavior.
func NewPlugin(optsFunc ...OptionFunc) api.Plugin {
	opts := newOptions()

	// Apply all provided option functions to configure the plugin
	for _, fn := range optsFunc {
		fn(opts)
	}

	if opts.jsExecutor == nil {
		panic("jsExecutor is required, please set it using WithJsExecutor()")
	}

	return api.Plugin{
		Name: opts.name,
		Setup: func(build api.PluginBuild) {
			// Normalize and validate esbuild options
			normalizeEsbuildOptions(build.InitialOptions)

			// Start processor chain - executed before build starts
			build.OnStart(func() (api.OnStartResult, error) {
				// Execute all start processors
				for _, processor := range opts.onStartProcessors {
					if err := processor(build.InitialOptions); err != nil {
						return api.OnStartResult{}, err
					}
				}
				return api.OnStartResult{}, nil
			})

			// Register all handlers for Vue, Sass, and HTML
			setupVueHandler(opts, &build)
			setupSassHandler(opts, &build)
			setupHtmlHandler(opts, &build)

			// End processor chain - executed after all processing is done
			build.OnEnd(func(result *api.BuildResult) (api.OnEndResult, error) {
				for _, processor := range opts.onEndProcessors {
					if err := processor(result, build.InitialOptions); err != nil {
						return api.OnEndResult{}, err
					}
				}
				return api.OnEndResult{}, nil
			})

			// Dispose processor chain - cleanup after build
			build.OnDispose(func() {
				for _, processor := range opts.onDisposeProcessors {
					processor(build.InitialOptions)
				}
			})
		},
	}
}
