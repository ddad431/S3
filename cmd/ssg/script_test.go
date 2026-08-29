package main

import (
	"os"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"ssg": main,
	})
}

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",

		// 保存 ssg 命令的路径地址，方便之后屏蔽整个 $PATH 后仍能调用 ssg
		Setup: func(e *testscript.Env) error {
			path := e.Getenv("PATH")
			ssgDir := strings.Split(path, string(os.PathListSeparator))[0]

			e.Setenv("SSG_PATH", ssgDir)
			return nil
		},
	})
}
