package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInsertMetadataCreated(t *testing.T) {
	tests := []struct {
		name    string
		content string
		hasMeta bool
	}{
		{
			name:    "no front matter",
			content: "# Hello world",
			hasMeta: false,
		},
		{
			name:    "has front matter",
			content: "---\ntitle: test\n---\n\n# Hello world",
			hasMeta: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := filepath.Join(os.TempDir(), "test.md")
			if err := os.WriteFile(file, []byte(tt.content), 0644); err != nil {
				t.Fatalf("create test file fail: %v", err)
			}

			if err := insertMetadataCreated(file, tt.hasMeta); err != nil {
				t.Fatalf("run insertCreatedMetadata() err: %v", err)
			}

			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read test file fail: %v", err)
			}

			var got = string(content)
			var want string
			if !tt.hasMeta {
				want = fmt.Sprintf("---\ncreated: %s\n---\n\n# Hello world", time.Now().Format("2006/01/02"))
			} else {
				want = fmt.Sprintf("---\ntitle: test\ncreated: %s\n---\n\n# Hello world", time.Now().Format("2006/01/02"))
			}

			if got != want {
				t.Errorf("\nGot:\n%s\nWant:\n%s", got, want)
			}
		})
	}
}

func TestGetPageData(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		content   string
		wantTitle string
		wantHTML  template.HTML
	}{
		{
			name:      "no front matter",
			filename:  "test.md",
			content:   "# Hello world",
			wantTitle: "test",
			wantHTML:  "<h1>Hello world</h1>\n",
		},
		{
			name:      "with front matter",
			filename:  "test.md",
			content:   "---\ntitle: Hello world\n---\n\n# Heading 1",
			wantTitle: "Hello world",
			wantHTML:  "<h1>Heading 1</h1>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := os.TempDir()
			src := filepath.Join(tmp, tt.filename)

			if err := os.WriteFile(src, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			data, err := getPageData(src)
			if err != nil {
				t.Fatalf("getPageData() error: %v", err)
			}

			if data.Title != tt.wantTitle {
				t.Errorf("Title\nGot\t%s\nWant\t%s\n", data.Title, tt.wantTitle)
			}

			if data.Content != tt.wantHTML {
				t.Errorf("HTML\nGot\t%s\nWant\t%s\n", data.Content, tt.wantHTML)
			}
		})
	}
}

func TestBuildPage(t *testing.T) {
	tmp := t.TempDir()
	name := "filename"
	ipath := filepath.Join(tmp, name+".md")
	opath := filepath.Join(tmp, "post.html")
	tpl := filepath.Join(tmp, "template.html")
	pcnt := "# Hello World"
	tcnt := `<!DOCTYPE html>
<html>
<head>
    <title>{{ .Title }}</title>
</head>
<body>
    <img src="./assets/logo1.png" />
    <img src="assets/logo2.png" />
    {{ .Content }}
</body>
</html>`
	want := `<!DOCTYPE html>
<html>
<head>
    <title>filename</title>
</head>
<body>
    <img src="../../assets/logo1.png" />
    <img src="../../assets/logo2.png" />
    <h1>Hello World</h1>

</body>
</html>`

	if err := os.WriteFile(ipath, []byte(pcnt), 0644); err != nil {
		t.Fatalf("create file 'post.md' fail: %v", err)
	}

	if err := os.WriteFile(tpl, []byte(tcnt), 0644); err != nil {
		t.Fatalf("create file 'template.html' fail: %v", err)
	}

	err := buildPage(tpl, ipath, opath)
	if err != nil {
		t.Fatalf("exec function 'buildPage' fail: %v", err)
	}

	got, err := os.ReadFile(opath)
	if err != nil {
		t.Fatalf("read generated html fail: %v", err)
	}

	if string(got) != want {
		t.Errorf("failed\nGot:\n%s\nWant:\n%s", string(got), want)
	}

}
