package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pvinchon/agent/internal/assistant"
	"github.com/pvinchon/agent/internal/git"
)

func main() {
	println("Hello, World!")

	println(git.BranchDefault())
	println(git.BranchCurrent())

	println(git.DiffWithDefault())

	resolveAssistant := assistant.Flag()
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
}
