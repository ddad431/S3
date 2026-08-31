package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

type Converter struct {
	parser   parser.Parser
	renderer html.Renderer
}

var converter Converter

func init() {
	converter = Converter{
		parser: parser.New(
			parser.WithExtensions(
				extension.GFMParser,
			),
		),
		renderer: html.New(
			html.WithUnsafe(),
		),
	}
}

func MDToHTML(source string) (template.HTML, error) {
    var content []byte
    var buf bytes.Buffer

    content, err := os.ReadFile(source);
    if err != nil {
        return "", fmt.Errorf("read '%s' fail: %w", source, err)
    }

    if err := converter.renderer.Render(&buf, content, converter.parser.Parse(content)); err != nil {
        return "", fmt.Errorf("render markdown fail: %w", err)
    }

    return template.HTML(buf.String()), nil
}
