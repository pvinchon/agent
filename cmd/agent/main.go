package main

import "github.com/pvinchon/agent/internal/git"

func main() {
	println("Hello, World!")

	println(git.BranchDefault())
	println(git.BranchCurrent())
}
