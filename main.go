package main

/*
	@author: Aviral Nigam
	@github: https://github.com/iAviPro
	@date: 15 Jun, 2020
*/

import (
	"fmt"
	"os"

	"github.com/ConsulScale/src" // Corrected import path
)

// Test code
func main() {
	// Original logic for arg check can remain, or be simplified if ExecuteGoConsulKVFunc handles it.
	// For now, keeping original logic but calling the Func variable.
	if len(os.Args) < 2 { // Simplified check: at least one command is expected.
		// ExecuteGoConsulKVFunc itself handles the no-command case by printing usage.
		// So, this specific check might be redundant if ExecuteGoConsulKVFunc covers it.
		// Let's defer to ExecuteGoConsulKVFunc for arg parsing and errors.
		// The original check `len(os.Args) <= 2 && os.Args[1] != "backup"` was complex
		// and might prevent ExecuteGoConsulKVFunc from even being called with its default "no args" behavior.
		// For tests to reliably mock ExecuteGoConsulKVFunc, main should just call it.
	}
	// Always call ExecuteGoConsulKVFunc; it will parse os.Args itself.
	src.ExecuteGoConsulKVFunc()
}
