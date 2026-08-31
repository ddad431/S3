package main

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
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
	content, err := MDToHTML(source)
	if err != nil {
		return pageData, err
	}
	pageData.Content = content

	// Title
	pageData.Title = strings.TrimSuffix(filepath.Base(source), ".md")

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
