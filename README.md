# esbuild-plugin-vue-go
English | [简体中文](README_zh-cn.md)

A pure Golang esbuild plugin for resolving and loading Vue3 single-file components (SFC), with no Node.js dependency.

This project is inspired by [pipe01/esbuild-plugin-vue3](https://github.com/pipe01/esbuild-plugin-vue3).


## Features & Limitations

> **This project is experimental and currently supports only limited features, Some advanced Vue SFC features or edge cases may not be fully supported.**

1. Supports standard Vue `<script>` and `<script setup>` blocks written in JavaScript or TypeScript.
2. `<template>` only supports standard Vue template syntax. Other template languages (such as Pug) are **not** supported.
3. `<style>` supports CSS, SCSS, and SASS. **Only relative path imports** are supported in Sass/SCSS.
4. Supports generating HTML files and automatic injection of built JS/CSS assets.  
   You can use a custom `IndexHtmlProcessor` to modify the HTML generation logic
5. Provides plugin hooks for custom processors at various build stages for advanced customization, including:`OnStartProcessor`/`OnVueResolveProcessor`/`OnVueLoadProcessor`/ `OnSassLoadProcessor`/`OnEndProcessor`/`OnDisposeProcessor`/`IndexHtmlProcessor`


## Quick Start
### 1. Install

```bash
go get github.com/buke/esbuild-plugin-vue-go
```

### 2. Example Usage

```go
import (
    "fmt"
    vueplugin "github.com/buke/esbuild-plugin-vue-go"
    "github.com/buke/esbuild-plugin-vue-go/engine"
    jsexecutor "github.com/buke/js-executor"
    "github.com/evanw/esbuild/pkg/api"
)

func main() {
    // 1. Create a new JavaScript executor and start it.
    jsExec, _ := jsexecutor.NewExecutor(
        jsexecutor.WithJsEngine(engine.NewVueCompilerFactory()),
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
}
```

### 3. Advanced Usage

You can use custom processors (hooks) to extend or modify the build process at various stages.  
For example, you can preprocess Vue or Sass files, customize HTML output, or run custom logic before/after build.

Below is an example of using custom processors:

```go
import (
    vueplugin "github.com/buke/esbuild-plugin-vue-go"
    "github.com/evanw/esbuild/pkg/api"
    "fmt"
)

// Custom processor: print a message before build starts
func myStartProcessor(buildOptions *api.BuildOptions) error {
    fmt.Println("Build is starting!")
    return nil
}

// Custom processor: modify Vue file content before compilation
func myVueLoadProcessor(content string, args api.OnLoadArgs, buildOptions *api.BuildOptions) (string, error) {
    // For example, inject a comment at the top of every Vue file
    return "// Processed by myVueLoadProcessor\n" + content, nil
}

func main() {
    // ... (create jsExec as in previous examples)

    vuePlugin := vueplugin.NewPlugin(
        vueplugin.WithJsExecutor(jsExec),
        vueplugin.WithOnStartProcessor(myStartProcessor),
        vueplugin.WithOnVueLoadProcessor(myVueLoadProcessor),
        // You can add other processors as needed
    )

    result := api.Build(api.BuildOptions{
        // ...esbuild options...
        Plugins: []api.Plugin{vuePlugin},
    })
}
```

**Supported Processors:**

- `OnStartProcessor`: Run custom logic before the build starts.
- `OnVueResolveProcessor`: Customize how Vue files are resolved.
- `OnVueLoadProcessor`: Transform or preprocess Vue file content before compilation.
- `OnSassLoadProcessor`: Transform or preprocess Sass/SCSS file content before compilation.
- `OnEndProcessor`: Run custom logic after the build finishes.
- `OnDisposeProcessor`: Cleanup logic after the build is disposed.
- `IndexHtmlProcessor`: Customize HTML file processing and asset injection after build.

## How It Works

This project uses [github.com/buke/js-executor](https://github.com/buke/js-executor) and [github.com/buke/quickjs-go](https://github.com/buke/quickjs-go) to embed a JavaScript engine (QuickJS) and run JavaScript in Go.  
It loads the official [@vue/compiler-sfc](https://www.npmjs.com/package/@vue/compiler-sfc) JavaScript code and calls it directly to compile `.vue` files.  
For Sass/SCSS support, it uses the pure JavaScript [Dart Sass package](https://www.npmjs.com/package/sass) (not the native binary), so all Sass/SCSS compilation is also handled inside the embedded JS engine—no Node.js or native binaries required.

**Therefore, you can compile Vue files without a Node.js runtime environment.**  
However, you still need to use `npm install` (or another package manager) to install the required JavaScript dependencies for your project, so that esbuild can find and build them.

## License

[Apache-2.0](LICENSE)
