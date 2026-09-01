package main

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type PageData struct {
	Title   string
	Content template.HTML
}

func buildPage(tpl, ipath, opath string) error {
	pageData, err := getPageData(ipath)
	if err != nil {
		return err
	}

	if err := renderTemplateToPage(tpl, pageData, opath); err != nil {
		return err
	}

	return nil
}

func getPageData(source string) (PageData, error) {
	var pageData PageData

	// Content
	doc, err := converter.convert(source)
	if err != nil {
		return pageData, err
	}
	pageData.Content = doc.HTML

	// Meta
	types := reflect.TypeOf(pageData)
	values := reflect.ValueOf(&pageData).Elem()
	for mk, mv := range doc.Meta {
		for i := range types.NumField() {
			if strings.EqualFold(mk, types.Field(i).Name) {
				mv, ok := mv.(string)
				if !ok {
					mv = ""
				}

				values.Field(i).Set(reflect.ValueOf(mv))
			}
		}
	}

	if pageData.Title == "" {
		pageData.Title = strings.TrimSuffix(filepath.Base(source), ".md")
	}

	return pageData, nil
}

func renderTemplateToPage(src string, data PageData, dst string) error {
	rule := strings.NewReplacer(
		`="./assets/`, `="../../assets/`,
		`="assets/`, `="../../assets/`,
	)

	tpl, err := template.ParseFiles(src)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}

	content := rule.Replace(buf.String())
	if err := os.WriteFile(dst, []byte(content), 0644); err != nil {
		return err
	}

	return nil
}
