package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/pvinchon/agent/internal/assistant"
	"github.com/pvinchon/agent/internal/fixer"
	"github.com/pvinchon/agent/internal/git"
	"github.com/pvinchon/agent/internal/reviewer"
	xlog "github.com/pvinchon/agent/internal/x/log"
)

func main() {
	println("Hello, World!")

	resolveLog := xlog.Flag()
	resolveAssistant := assistant.Flag()
	resolveReviewers := reviewer.Flag()
	flag.Parse()

	slog.SetDefault(resolveLog())

	assistant, err := resolveAssistant()
	if err != nil {
		log.Fatal(err)
	}

	reviewers, err := resolveReviewers()
	if err != nil {
		log.Fatal(err)
	}

	diff, err := git.DiffWithDefault()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Reviewing difference")
	issues, errs := reviewer.Review(reviewers, diff, assistant)
	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}

	log.Println("Fixing issues")
	if err := fixer.Fix(issues, diff, assistant); err != nil {
		log.Fatal(err)
	}
}
