package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pvinchon/agent/internal/assistant"
	"github.com/pvinchon/agent/internal/git"
	"github.com/pvinchon/agent/internal/reviewer"
)

func main() {
	println("Hello, World!")

	println(git.BranchDefault())
	println(git.BranchCurrent())

	println(git.DiffWithDefault())

	resolveAssistant := assistant.Flag()
	resolveReviewers := reviewer.Flag()
	flag.Parse()

	a, err := resolveAssistant()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	joke, err := assistant.Prompt(a, "Tell me a joke.")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(joke)

	diff, err := git.DiffWithDefault()
	if err != nil {
		log.Fatal(err)
	}

	reviewers, err := resolveReviewers()
	if err != nil {
		log.Fatal(err)
	}

	issues, errs := reviewer.Review(reviewers, diff, a)
	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
	for _, f := range issues {
		fmt.Printf("Issue: %s\n", f.Title)
		fmt.Printf("Description: %s\n", f.Description)
		fmt.Printf("Location: %s\n", f.Location)
		fmt.Printf("Severity: %s\n", f.Severity)
		fmt.Println()
	}
}
