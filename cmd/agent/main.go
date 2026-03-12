package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage: agent <command> [flags]

Commands:
  review  Run diff and review; outputs issues as JSON to stdout
  fix     Read issues as JSON from stdin and apply fixes
  loop    Run review and fix in a loop until no issues or max attempts
  help    Show usage for a command

`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd, args := os.Args[1], os.Args[2:]

	switch cmd {
	case "review":
		runReview(args, os.Stdout)
	case "fix":
		runFix(args, os.Stdin)
	case "loop":
		runLoop(args)
	case "help":
		runHelp(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}
}

func runHelp(args []string) {
	if len(args) == 0 {
		usage()
		return
	}
	switch args[0] {
	case "review":
		fs := flag.NewFlagSet("review", flag.ContinueOnError)
		reviewFlags(fs)
		reviewUsage(fs)
	case "fix":
		fs := flag.NewFlagSet("fix", flag.ContinueOnError)
		fixFlags(fs)
		fixUsage(fs)
	case "loop":
		fs := flag.NewFlagSet("loop", flag.ContinueOnError)
		loopFlags(fs)
		loopUsage(fs)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		usage()
		os.Exit(1)
	}
}
