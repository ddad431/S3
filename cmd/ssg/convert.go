package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	meta "github.com/yuin/goldmark-meta/v2"
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

type Document struct {
	HTML template.HTML
	Meta map[string]any
}

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
				meta.Parser,
			),
		),
		renderer: html.New(
			html.WithUnsafe(),
		),
	}
}

func (c Converter) convert(source string) (Document, error) {
	var buf bytes.Buffer

	content, err := os.ReadFile(source)
	if err != nil {
		return Document{}, fmt.Errorf("read '%s' fail: %w", source, err)
	}

	doc := c.parser.Parse(content)
	if err := c.renderer.Render(&buf, content, doc); err != nil {
		return Document{}, fmt.Errorf("render markdown fail: %w", err)
	}

	return Document{
		HTML: template.HTML(buf.String()),
		Meta: doc.(*ast.Document).Metadata(),
	}, nil
}
