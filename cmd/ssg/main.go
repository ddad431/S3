package main

import (
	"flag"
	"fmt"
	"os"
)

type rootOpts struct {
	version bool
}

func newRootCmd(opts *rootOpts) *flag.FlagSet {
	cmd := flag.NewFlagSet("ssg", flag.ExitOnError)

	cmd.BoolVar(&opts.version, "v", false, "print version")
	cmd.BoolVar(&opts.version, "version", false, "print version")

	cmd.Usage = func() {
		fmt.Println("usage: ssg [-v|--version] [-h|--help]")
		fmt.Println("           <command> [<args>]")
		fmt.Println()
		fmt.Println("command")
		fmt.Println("  doctor    check dependencies and directory layout")
		fmt.Println("  init      initialize directory layout")
		fmt.Println("  build     build static page")
		fmt.Println()
	}

	return cmd
}

func main() {
	var opts rootOpts

	cmd := newRootCmd(&opts)
	if err := cmd.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		cmd.Usage()
		os.Exit(1)
	}

	if opts.version {
		fmt.Printf("ssg version 0.1.0")
		return
	}

	subCmd := os.Args[1]
	subArgs := os.Args[2:]
	switch subCmd {
	case "doctor":
		os.Exit(runDoctor(subArgs))

	case "init":
		os.Exit(runInit(subArgs))

	case "build":
		os.Exit(runBuild(subArgs))

	default:
		fmt.Printf("unknown command: %v", os.Args[1])
		return
	}
}
