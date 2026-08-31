package main

import (
	"embed"
	"os"
	"path/filepath"
)

//go:embed templates/*
var templates embed.FS

func createTemplateFile(path string) error {
	data, err := templates.ReadFile("templates/" + filepath.Base(path))
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	return nil
}
