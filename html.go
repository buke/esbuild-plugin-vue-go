// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package vueplugin

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/evanw/esbuild/pkg/api"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

// HtmlProcessorOptions holds builder functions for script and CSS tag attributes.
type HtmlProcessorOptions struct {
	ScriptAttrBuilder func(filename string, htmlSourceFile string) []html.Attribute // JS script tag attribute builder
	CssAttrBuilder    func(filename string, htmlSourceFile string) []html.Attribute // CSS link tag attribute builder
}

// NewHtmlProcessor returns an IndexHtmlProcessor that injects JS and CSS tags and removes specified nodes.
func NewHtmlProcessor(htmlProcessorOptions HtmlProcessorOptions) IndexHtmlProcessor {
	return func(doc *html.Node, result *api.BuildResult, opts *options, build *api.PluginBuild) error {
		if htmlProcessorOptions.ScriptAttrBuilder == nil {
			// Default JS script tag attribute builder
			htmlProcessorOptions.ScriptAttrBuilder = func(filename string, htmlFile string) []html.Attribute {
				// Compute relative path between filename and htmlSourceFile
				relPath, _ := filepath.Rel(filepath.Dir(htmlFile), filename)
				if relPath != "" {
					filename = filepath.ToSlash(relPath) // Ensure forward slashes for web compatibility
				}
				return []html.Attribute{
					{Key: "crossorigin", Val: ""},
					{Key: "type", Val: "module"},
					{Key: "src", Val: filename},
				}
			}
		}

		if htmlProcessorOptions.CssAttrBuilder == nil {
			// Default CSS link tag attribute builder
			htmlProcessorOptions.CssAttrBuilder = func(filename string, htmlFile string) []html.Attribute {
				relPath, _ := filepath.Rel(filepath.Dir(htmlFile), filename)
				if relPath != "" {
					filename = filepath.ToSlash(relPath) // Ensure forward slashes for web compatibility
				}
				return []html.Attribute{
					{Key: "crossorigin", Val: ""},
					{Key: "rel", Val: "stylesheet"},
					{Key: "href", Val: filename},
				}
			}
		}

		htmlFile, _ := filepath.Abs(opts.indexHtmlOptions.OutFile)

		// Find <head> tag in the HTML document
		headNode := htmlquery.FindOne(doc, "//head")
		// Process all output files
		for _, outputFile := range result.OutputFiles {
			// Normalize output file path
			outputFile, _ := filepath.Abs(outputFile.Path)

			// Fixed file filtering logic
			shouldInclude := false
			for _, entryPoint := range build.InitialOptions.EntryPoints {
				entry := filepath.Base(entryPoint)
				entryPointPrefix := strings.TrimSuffix(entry, filepath.Ext(entry))
				if strings.HasPrefix(filepath.Base(outputFile), entryPointPrefix) {
					shouldInclude = true
					break // Found a match, no need to check other entry points
				}
			}
			if !shouldInclude {
				continue // Skip files not generated from entry points
			}

			// Add tags based on file extension
			switch filepath.Ext(outputFile) {
			case ".js":
				// Add JavaScript file
				scriptNode := &html.Node{
					Type: html.ElementNode,
					Data: "script",
					Attr: htmlProcessorOptions.ScriptAttrBuilder(outputFile, htmlFile),
				}
				headNode.AppendChild(scriptNode)
				newline := &html.Node{
					Type: html.TextNode,
					Data: "\n",
				}
				headNode.AppendChild(newline)
			case ".css":
				// Add CSS file
				linkNode := &html.Node{
					Type: html.ElementNode,
					Data: "link",
					Attr: htmlProcessorOptions.CssAttrBuilder(outputFile, htmlFile),
				}
				headNode.AppendChild(linkNode)
				newline := &html.Node{
					Type: html.TextNode,
					Data: "\n",
				}
				headNode.AppendChild(newline)
			}
		}

		// Remove specified HTML nodes by XPath
		for _, xpath := range opts.indexHtmlOptions.RemoveTagXPaths {
			nodes := htmlquery.Find(doc, xpath)
			for _, node := range nodes {
				if node.Parent != nil {
					node.Parent.RemoveChild(node)
				}
			}
		}

		return nil
	}
}

// htmlHandle 处理HTML文件
func setupHtmlHandler(opts *options, build *api.PluginBuild) {
	build.OnEnd(func(result *api.BuildResult) (api.OnEndResult, error) {
		// 跳过不需要处理HTML的情况
		if opts.indexHtmlOptions.SourceFile == "" || !build.InitialOptions.Write {
			return api.OnEndResult{}, nil
		}
		if result.Metafile == "" {
			return api.OnEndResult{}, fmt.Errorf("metafile is nil")
		}
		if opts.indexHtmlOptions.OutFile == "" || opts.indexHtmlOptions.SourceFile == "" {
			return api.OnEndResult{}, fmt.Errorf("outFile or sourceFile is nil")
		}

		if len(opts.indexHtmlOptions.IndexHtmlProcessors) == 0 {
			// 如果没有自定义处理器，添加默认的HTML处理器
			opts.indexHtmlOptions.IndexHtmlProcessors = []IndexHtmlProcessor{
				NewHtmlProcessor(HtmlProcessorOptions{}),
			}
		}

		// 读取并解析HTML文件
		sourceFile, err := os.Open(opts.indexHtmlOptions.SourceFile)
		if err != nil {
			return api.OnEndResult{}, fmt.Errorf("failed to open source file: %v", err)
		}
		defer sourceFile.Close()

		utf8Reader, err := detectAndConvertToUTF8(sourceFile)
		if err != nil {
			return api.OnEndResult{}, fmt.Errorf("failed to convert source file to UTF-8: %v", err)
		}

		doc, _ := htmlquery.Parse(utf8Reader)
		// 执行HTML处理器链
		if len(opts.indexHtmlOptions.IndexHtmlProcessors) > 0 {
			for _, processor := range opts.indexHtmlOptions.IndexHtmlProcessors {
				if err := processor(doc, result, opts, build); err != nil {
					return api.OnEndResult{}, err
				}
			}
		}

		// 渲染并保存HTML
		var buf bytes.Buffer
		err = html.Render(&buf, doc)
		if err != nil {
			return api.OnEndResult{}, err
		}

		err = os.WriteFile(opts.indexHtmlOptions.OutFile, buf.Bytes(), 0644)
		if err != nil {
			return api.OnEndResult{}, err
		}

		return api.OnEndResult{}, nil
	})
}

func detectAndConvertToUTF8(r io.Reader) (io.Reader, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	encoding, _, _ := charset.DetermineEncoding(b, "")

	utf8Reader := transform.NewReader(bytes.NewReader(b), encoding.NewDecoder())
	return utf8Reader, nil
}
