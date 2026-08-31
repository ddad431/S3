package main

import (
	"os"
	"path/filepath"
	"testing"
)

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
