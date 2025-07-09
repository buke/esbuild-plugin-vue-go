// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package vueplugin_test

import (
	"fmt"

	vueplugin "github.com/buke/esbuild-plugin-vue-go"
	qjscompiler "github.com/buke/esbuild-plugin-vue-go/engines/quickjs-go"
	jsexecutor "github.com/buke/js-executor"
	"github.com/evanw/esbuild/pkg/api"
)

// Example demonstrates how to use the Vue plugin with esbuild.
// This function shows the typical workflow: create a JS executor, configure the plugin, and run a build.
func Example() {
	// 1. Create a new JavaScript executor and start it.
	jsExec, _ := jsexecutor.NewExecutor(
		jsexecutor.WithJsEngine(qjscompiler.NewVueCompilerFactory()),
	)
	if err := jsExec.Start(); err != nil {
		panic(err)
	}
	defer jsExec.Stop()

	// 2. Create a Vue plugin with the JavaScript executor and additional options.
	vuePlugin := vueplugin.NewPlugin(
		vueplugin.WithJsExecutor(jsExec),
		vueplugin.WithIndexHtmlOptions(vueplugin.IndexHtmlOptions{
			SourceFile:      "example/vue-example/index.html",
			OutFile:         "example/dist/index.html",
			RemoveTagXPaths: []string{"//script[@src='/src/main.ts']"},
		}),
		vueplugin.WithOnEndProcessor(vueplugin.SimpleCopy(map[string]string{
			"example/vue-example/public/favicon.ico": "example/dist/favicon.ico",
		})),
	)

	// 3. Build the Vue application using esbuild with the Vue plugin.
	buildResult := api.Build(api.BuildOptions{
		EntryPoints: []string{"example/vue-example/src/main.ts"},
		PublicPath:  "example/dist/assets",
		Tsconfig:    "example/vue-example/tsconfig.app.json",
		Loader: map[string]api.Loader{
			".png":  api.LoaderFile,
			".scss": api.LoaderCSS,
			".sass": api.LoaderCSS,
			".svg":  api.LoaderFile,
		},
		Target:            api.ES2020,
		Platform:          api.PlatformBrowser,
		Format:            api.FormatESModule,
		Bundle:            true,
		Outdir:            "example/dist/assets",
		Plugins:           []api.Plugin{vuePlugin},
		Metafile:          true,
		Write:             true,
		EntryNames:        "[dir]/[name]-[hash]",
		Splitting:         true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
	})

	// 4. Print build result or errors.
	if buildResult.Errors != nil {
		for _, err := range buildResult.Errors {
			fmt.Printf("Build error: %s\n", err.Text)
		}
	} else {
		fmt.Println("Build succeeded!")
	}

	// Output:
	// Build succeeded!

}
