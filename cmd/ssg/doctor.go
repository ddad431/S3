package main

import (
	"flag"
	"fmt"
	"strings"
)

// $ ssg doctor
// directory layout
//         required files exist
//
// $ ssg doctor
//
// directory layout
//         missing required files or permission denied
//
//         • posts/                   (directory)
//         • index.md
//         • themes/default/index.html
//         • themes/default/post.html
//
//         Run 'ssg init' to fix missing files
//
// found 1 issue

func newDoctorCmd() *flag.FlagSet {
	cmd := flag.NewFlagSet("doctor", flag.ExitOnError)

	cmd.Usage = func() {
		fmt.Println("usage: ssg doctor")
	}

	return cmd
}

func runDoctor(args []string) int {
	cmd := newDoctorCmd()
	if err := cmd.Parse(args); err != nil {
		return 1
	}

	doCheck()
	return 0
}

type CheckResult struct {
	Name    string
	Ok      bool
	Message string
	Detail  string
}

func doCheck() {
	results := []CheckResult{
		checkDirectoryLayout(),
	}

	var issues = 0
	for _, res := range results {
		if !res.Ok {
			issues++
			fmt.Printf("%s\n\t%s\n%s\n", res.Name, res.Message, res.Detail)
			continue
		}

		// NOTE 当 detail 为空时，message 后的空格在测试时很难发现
		if res.Detail == "" {
			fmt.Printf("%s\n\t%s%s\n\n", res.Name, res.Message, res.Detail)
		} else {
			fmt.Printf("%s\n\t%s %s\n\n", res.Name, res.Message, res.Detail)
		}
	}

	switch issues {
	case 0:
		fmt.Printf("")
	case 1:
		fmt.Printf("found %v issue\n", issues)
	default:
		fmt.Printf("found %v issues\n", issues)
	}
}

func checkDirectoryLayout() CheckResult {
	const (
		ProblemMissingDir   = "(directory)"
		ProblemMissingFile  = ""
		ProblemExpectedDir  = "(expected directory, get file)"
		ProblemExpectedFile = "(expected file, get directory)"
	)

	type PathProblem struct {
		Path    string
		Message string
	}

	var problems []PathProblem
	for _, path := range Required {
		state := CheckPath(path.Path, path.Type)

		switch state.State {
		case PathStateOk:

		case PathStateMissing:
			if state.Type == PathDir {
				problems = append(problems, PathProblem{Path: path.Path, Message: ProblemMissingDir})
			} else if state.Type == PathFile {
				problems = append(problems, PathProblem{Path: path.Path, Message: ProblemMissingFile})
			}

		case PathStateConflict:
			if state.ExpectedType == PathDir {
				problems = append(problems, PathProblem{Path: path.Path, Message: ProblemExpectedDir})
			} else if state.ExpectedType == PathFile {
				problems = append(problems, PathProblem{Path: path.Path, Message: ProblemExpectedFile})
			}

		case PathStateError:
			problems = append(problems, PathProblem{Path: path.Path, Message: "(" + state.Err.Error() + ")"})
		}
	}

	var maxlen int
	for _, problem := range problems {
		pathlen := len(problem.Path)
		if pathlen > maxlen {
			maxlen = pathlen
		}
	}

	if len(problems) > 0 {
		var list []string
		for _, path := range problems {
			list = append(list, "\t• "+fmt.Sprintf("%-*s", maxlen, path.Path)+" "+path.Message+"\n")
		}

		info := []string{"\n"}
		info = append(info, list...)
		info = append(info, "\n\tRun 'ssg init' to fix missing files\n")

		return CheckResult{
			Name:    "directory layout",
			Ok:      false,
			Message: "missing required files or permission denied",
			Detail:  strings.Join(info, ""),
		}
	}

	return CheckResult{
		Name:    "directory layout",
		Ok:      true,
		Message: "required files exist",
	}
}
