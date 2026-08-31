package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// $ ssg init
// initialized
//
// $ ssg init
//	created   posts
//	created   index.md
//	created   themes\default\index.html
//	created   themes\default\post.html
//
// initialization successful
//
// $ mkdir index.md
// $ ssg init
//	created   posts
//
// initialization failed: 'index.md' is a directory, expected file (use --force/-f to overwrite)
//
// $ mkdir -p {themes/default,posts,index.md} && touch themes/default/{index,post}.html
// $ ssg init
// initialization failed: 'index.md' is a directory, expected file (use --force/-f to overwrite)
//
// $ ssg init --force
// initialized
//
// $ ssg init --force
//	created   posts
//	created   index.md
//	created   themes\default\index.html
//	created   themes\default\post.html
//
// $ mkdir index.md
// $ ssg init --force
//	created   posts
//	modified  index.md (directory -> file)
//	created   themes\default\index.html
//	created   themes\default\post.html
//
// initialization successful
//
// $ mkdir -p {themes/default,posts,index.md} && touch themes/default/{index,post}.html
// $ ssg init --force
//  modified  index.md (directory -> file)
//
// initialization successful

type initOpts struct {
	path  string
	force bool
}

func (opts *initOpts) validate() error {
	vpath, err := ValidatePath(opts.path, PathDir)
	if err != nil {
		return err
	}
	opts.path = vpath

	return nil
}

func newInitCmd(opts *initOpts) *flag.FlagSet {
	cmd := flag.NewFlagSet("init", flag.ExitOnError)

	cmd.StringVar(&opts.path, "p", "", "target directory path")
	cmd.StringVar(&opts.path, "path", "", "target directory path")
	cmd.BoolVar(&opts.force, "f", false, "force overwrite on conflict")
	cmd.BoolVar(&opts.force, "force", false, "force overwrite on conflict")

	cmd.Usage = func() {
		fmt.Println("usage: ssg init [-p|--path <path>] [-f|--force]")
		fmt.Println("")
		fmt.Println("  -f, --force")
		fmt.Println("        force overwrite on conflict")
		fmt.Println("  -p, --path <path>")
		fmt.Println("        target directory path")
		fmt.Println()
	}

	return cmd
}

func runInit(args []string) int {
	var opts initOpts

	cmd := newInitCmd(&opts)
	if err := cmd.Parse(args); err != nil {
		return 1
	}

	if err := opts.validate(); err != nil {
		fmt.Printf("initialization failed: %s", err)
		return 1
	}

	if err := doInitialize(opts); err != nil {
		fmt.Printf("initialization failed: %s", err)
		return 1
	}

	return 0
}

func doInitialize(opts initOpts) error {
	var health = 0
	var printed = false

	states := checkLayout(opts.path)
	for _, state := range states {
		var err error

		switch state.State {
		case PathStateOk:
			health++
			continue

		case PathStateMissing:
			err = handlePathMissing(state)
			printed = true

		case PathStateConflict:
			err = handlePathConflict(state, opts.force)
			if opts.force {
				printed = true
			}

		case PathStateError:
			err = state.Err
		}

		if err != nil {
			if printed {
				fmt.Println()
			}
			return err
		}
	}

	if health == len(states) {
		fmt.Println("initialized")
		return nil
	}

	fmt.Println("\ninitialization successful")
	return nil
}

func handlePathMissing(state PathCheck) error {
	var err error

	switch state.Type {
	case PathDir:
		err = createDirectory(state.Path)
	case PathFile:
		err = createFile(state.Path)
	}

	if err == nil {
		fmt.Printf("  %-8s  %s\n", "created", state.Path)
	}
	return err
}

func handlePathConflict(state PathCheck, force bool) error {
	if !force {
		var msg string
		switch state.ExpectedType {
		case PathDir:
			msg = "is a file, expected directory"
		case PathFile:
			msg = "is a directory, expected file"
		}

		return fmt.Errorf("'%s' %s (use --force/-f to overwrite)\n", state.Path, msg)
	}

	if err := os.RemoveAll(state.Path); err != nil {
		return err
	}

	var err error

	switch state.ExpectedType {
	case PathDir:
		err = createDirectory(state.Path)
		if err == nil {
			fmt.Printf("  %-8s  %s (file -> directory)\n", "modified", state.Path)
		}
	case PathFile:
		err = createFile(state.Path)
		if err == nil {
			fmt.Printf("  %-8s  %s (directory -> file)\n", "modified", state.Path)
		}
	}

	return err
}

func checkLayout(path string) []PathCheck {
	var res []PathCheck
	for _, rpath := range Required {
		res = append(res, CheckPath(filepath.Join(path, rpath.Path), rpath.Type))
	}

	return res
}

func createFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if strings.Contains(path, "themes/default/") || strings.Contains(path, "themes\\default\\") {
		if err := createTemplateFile(path); err != nil {
			return err
		}
		return nil
	}

	if err := os.WriteFile(path, nil, 0644); err != nil {
		return err
	}

	return nil
}

func createDirectory(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	return nil
}
