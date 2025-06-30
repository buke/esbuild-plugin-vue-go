# esbuild-plugin-vue-go
[English](README.md) | 简体中文

[![Test](https://github.com/buke/esbuild-plugin-vue-go/workflows/Test/badge.svg)](https://github.com/buke/esbuild-plugin-vue-go/actions?query=workflow%3ATest)
[![codecov](https://codecov.io/gh/buke/esbuild-plugin-vue-go/graph/badge.svg?token=sCKbIlGJE3)](https://codecov.io/gh/buke/esbuild-plugin-vue-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/buke/esbuild-plugin-vue-go)](https://goreportcard.com/report/github.com/buke/esbuild-plugin-vue-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/buke/esbuild-plugin-vue-go?status.svg)](https://pkg.go.dev/github.com/buke/esbuild-plugin-vue-go?tab=doc)

一个纯 Golang 的 esbuild 插件，用于解析和加载 Vue3 单文件组件（SFC），无需 Node.js 依赖。

本项目受到 [pipe01/esbuild-plugin-vue3](https://github.com/pipe01/esbuild-plugin-vue3) 启发。

## 功能与限制

> **本项目为实验性实现，目前仅支持有限功能，部分高级 Vue SFC 特性或边缘场景可能暂不支持。**

1. 支持标准 Vue `<script>` 和 `<script setup>`，可使用 JavaScript 或 TypeScript 编写。
2. `<template>` 仅支持标准 Vue 模板语法，不支持其他模板语言（如 Pug）。
3. `<style>` 支持 CSS、SCSS 和 SASS，Sass/SCSS 中**仅支持相对路径引用**。
4. 支持生成 HTML 文件并自动注入构建后的 JS/CSS 资源。  
   你可以通过自定义 `IndexHtmlProcessor` 灵活修改 HTML 生成逻辑。
5. 提供插件钩子，可在各个构建阶段自定义处理流程，包括：  
   `OnStartProcessor`、`OnVueResolveProcessor`、`OnVueLoadProcessor`、`OnSassLoadProcessor`、`OnEndProcessor`、`OnDisposeProcessor`、`IndexHtmlProcessor`

## 快速开始

### 1. 安装

```bash
go get github.com/buke/esbuild-plugin-vue-go
```

### 2. 基本用法示例

```go
import (
    "fmt"
    vueplugin "github.com/buke/esbuild-plugin-vue-go"
    "github.com/buke/esbuild-plugin-vue-go/engine"
    jsexecutor "github.com/buke/js-executor"
    "github.com/evanw/esbuild/pkg/api"
)

func main() {
    // 1. 创建 JS 执行器并启动
    jsExec, _ := jsexecutor.NewExecutor(
        jsexecutor.WithJsEngine(engine.NewVueCompilerFactory()),
    )
    if err := jsExec.Start(); err != nil {
        panic(err)
    }
    defer jsExec.Stop()

    // 2. 创建 Vue 插件并配置相关选项
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

    // 3. 使用 esbuild 构建 Vue 应用
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

    // 4. 输出构建结果或错误
    if buildResult.Errors != nil {
        for _, err := range buildResult.Errors {
            fmt.Printf("Build error: %s\n", err.Text)
        }
    } else {
        fmt.Println("Build succeeded!")
    }
}
```

### 3. 进阶用法

你可以通过自定义 Processor（钩子）在不同构建阶段扩展或修改构建流程。  
例如，可以预处理 Vue 或 Sass 文件，自定义 HTML 输出，或在构建前后执行自定义逻辑。

以下是自定义 Processor 的示例：

```go
import (
    vueplugin "github.com/buke/esbuild-plugin-vue-go"
    "github.com/evanw/esbuild/pkg/api"
    "fmt"
)

// 构建开始前自定义处理
func myStartProcessor(buildOptions *api.BuildOptions) error {
    fmt.Println("Build is starting!")
    return nil
}

// 编译前自定义处理 Vue 文件内容
func myVueLoadProcessor(content string, args api.OnLoadArgs, buildOptions *api.BuildOptions) (string, error) {
    // 例如：在每个 Vue 文件顶部插入注释
    return "// Processed by myVueLoadProcessor\n" + content, nil
}

func main() {
    // ...（如前例创建 jsExec）

    vuePlugin := vueplugin.NewPlugin(
        vueplugin.WithJsExecutor(jsExec),
        vueplugin.WithOnStartProcessor(myStartProcessor),
        vueplugin.WithOnVueLoadProcessor(myVueLoadProcessor),
        // 还可以添加其他 Processor
    )

    result := api.Build(api.BuildOptions{
        // ...esbuild 相关配置...
        Plugins: []api.Plugin{vuePlugin},
    })
}
```

**支持的 Processor 列表：**

- `OnStartProcessor`：构建开始前执行自定义逻辑
- `OnVueResolveProcessor`：自定义 Vue 文件的解析方式
- `OnVueLoadProcessor`：编译前处理或转换 Vue 文件内容
- `OnSassLoadProcessor`：编译前处理或转换 Sass/SCSS 文件内容
- `OnEndProcessor`：构建结束后执行自定义逻辑
- `OnDisposeProcessor`：构建结束后清理资源
- `IndexHtmlProcessor`：自定义 HTML 文件处理和资源注入

## 工作原理

本项目通过 [github.com/buke/js-executor](https://github.com/buke/js-executor) 和 [github.com/buke/quickjs-go](https://github.com/buke/quickjs-go) 在 Go 中嵌入 JavaScript 引擎（QuickJS）并运行 JavaScript。  
加载官方 [@vue/compiler-sfc](https://www.npmjs.com/package/@vue/compiler-sfc) JavaScript 代码，直接调用其 API 编译 `.vue` 文件。  
Sass/SCSS 支持则通过纯 JavaScript 的 [Dart Sass 包](https://www.npmjs.com/package/sass)（非原生二进制），所有 Sass/SCSS 编译也在嵌入的 JS 引擎中完成，无需 Node.js 或原生依赖。

**因此，你可以在没有 Node.js 环境的情况下编译 Vue 文件。**  
但你仍需通过 `npm install`（或其他包管理工具）安装项目所需的 JavaScript 依赖，以便 esbuild 能正确找到并构建它们。

## 许可证

[Apache-2.0](LICENSE)