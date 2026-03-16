package main

import "testing"

// TestUsage is a smoke test that verifies usage() doesn't panic.
func TestUsage(t *testing.T) {
	usage()
}

// TestRunHelp_noArgs verifies that help with no arguments prints general usage
// without panicking.
func TestRunHelp_noArgs(t *testing.T) {
	runHelp([]string{})
}

// TestRunHelp_review verifies that "help review" prints review usage without
// panicking.
func TestRunHelp_review(t *testing.T) {
	runHelp([]string{"review"})
}

// TestRunHelp_fix verifies that "help fix" prints fix usage without panicking.
func TestRunHelp_fix(t *testing.T) {
	runHelp([]string{"fix"})
}

// TestRunHelp_loop verifies that "help loop" prints loop usage without
// panicking.
func TestRunHelp_loop(t *testing.T) {
	runHelp([]string{"loop"})
}
