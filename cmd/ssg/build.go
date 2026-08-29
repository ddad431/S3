package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type buildOpts struct {
	path string
}

func (opts *buildOpts) validate() error {
	vpath, err := ValidatePath(opts.path, PathDir)
	if err != nil {
		return err
	}
	opts.path = vpath

	return nil
}

func newBuildCmd(opts *buildOpts) *flag.FlagSet {
	cmd := flag.NewFlagSet("build", flag.ExitOnError)

	cmd.StringVar(&opts.path, "s", "", "path to source directory")
	cmd.StringVar(&opts.path, "source", "", "path to source directory")

	cmd.Usage = func() {
		fmt.Println("usage: ssg build [-s|--source <path>]")
		fmt.Println()
		fmt.Println("  -s, --source")
		fmt.Println("      path to source directory")
		fmt.Println()
	}

	return cmd
}

func runBuild(args []string) int {
	var opts buildOpts

	cmd := newBuildCmd(&opts)
	if err := cmd.Parse(args); err != nil {
		return 1
	}

	if err := opts.validate(); err != nil {
		fmt.Printf("build failed: %s\n", err)
		return 1
	}

	if err := doBuild(opts.path); err != nil {
		fmt.Printf("%s\n", err)
		return 1
	}

	return 0
}

func doBuild(source string) error {
	if err := precheck(source); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if err := preclean(source); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if err := buildIndex(source); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if err := buildPosts(source); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	return nil
}

func precheck(source string) error {
	states := checkLayout(source)

	for _, state := range states {
		switch state.State {
		case PathStateOk:
			continue

		case PathStateMissing:
			if state.Type == PathFile {
				return fmt.Errorf("file '%s' does not exist", state.Path)
			}
			if state.Type == PathDir {
				return fmt.Errorf("directory '%s' does not exist", state.Path)
			}

		case PathStateConflict:
			if state.Type == PathFile {
				return fmt.Errorf("'%s' is a file, expected a directory", state.Path)
			}
			if state.Type == PathDir {
				return fmt.Errorf("'%s' is a directory, expected a file", state.Path)
			}

		case PathStateError:
			return state.Err
		}
	}

	return nil
}

func preclean(source string) error {
	bpath := filepath.Join(source, "build")
	if err := os.RemoveAll(bpath); err != nil { // NOTE 当目录不存在时，RemoveAll 返回的是 nil 不是 error
		return fmt.Errorf("cannot clean build directory '%s': %w", bpath, err)
	}

	if err := os.MkdirAll(bpath, 0755); err != nil {
		return fmt.Errorf("cannot create build directory '%s': %w", bpath, err)
	}

	return nil
}

func buildIndex(source string) error {
	theme := "default"
	base := filepath.Join("themes", theme)
	icss := filepath.Join(base, "index.css")
	pcss := filepath.Join(base, "post.css")

	if err := os.MkdirAll(filepath.Join("build", base), 0755); err != nil {
		return err
	}

	files := []string{icss, pcss}
	for _, file := range files {
		if err := CopyFile(file, filepath.Join("build", file)); err != nil {
			return err
		}
	}

	src := filepath.Join(source, "index.md")
	dst := filepath.Join(source, "build", "index.html")
	if err := mdToHTML(src, dst, icss); err != nil {
		return err
	}

	return nil
}

func buildPosts(source string) error {
	path := filepath.Join(source, "posts")
	topics, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("'%s' not found: %w", path, err)
		}
		return fmt.Errorf("cannot access '%s': %w", path, err)
	}

	if len(topics) == 0 {
		return nil
	}

	for _, topic := range topics {
		if !topic.IsDir() {
			continue
		}
		rpath := filepath.Join("posts", topic.Name())

		if err := buildTopicImages(filepath.Join(source, rpath, "images"), filepath.Join(source, "build", rpath, "images")); err != nil {
			return err
		}

		if err := buildTopicPosts(filepath.Join(source, rpath), filepath.Join(source, "build", rpath)); err != nil {
			return err
		}
	}

	return nil
}

func buildTopicImages(src, dst string) error {
	imgs, err := os.ReadDir(src)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("cannot access '%s': %w", src, err)
	}

	if len(imgs) == 0 {
		return nil
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("cannot create directory '%s': %w", dst, err)
	}

	for _, img := range imgs {
		if img.IsDir() {
			continue
		}

		s := filepath.Join(src, img.Name())
		d := filepath.Join(dst, img.Name())
		if err := CopyFile(s, d); err != nil {
			return fmt.Errorf("cannot copy '%s' to '%s': %w", s, d, err)
		}
	}

	return nil
}

func buildTopicPosts(src, dst string) error {
	files, err := os.ReadDir(src)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("cannot access '%s': %w", src, err)
	}

	if len(files) == 0 {
		return nil
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("create '%s' fail: %w", dst, err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".md" {
			continue
		}

		ipath := filepath.Join(src, file.Name())
		opath := filepath.Join(dst, strings.TrimSuffix(file.Name(), ".md")+".html")
		cpath := filepath.Clean("../../themes/default/post.css")
		if err := mdToHTML(ipath, opath, cpath); err != nil {
			return err
		}
	}

	return nil
}

func mdToHTML(ipath, opath, cpath string) error {
	// pandoc ipath --from=markdown --to=html5 --css=cpath --standalone=true --output=opath
	cmd := exec.Command("pandoc",
		ipath,
		"--from=markdown",
		"--to=html5",
		"--css="+cpath,
		"--standalone=true",
		"--output="+opath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pandoc %s : %w", string(output), err)
	}

	return nil
}
