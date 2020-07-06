package main

/*
	@author: Aviral Nigam
	@github: https://github.com/iAviPro
	@date: 15 Jun, 2020
*/

import (
	"fmt"
	"os"

	"github.com/iAviPro/goConsulKV/src"
)

// Test code
func main() {
	if len(os.Args) <= 2 && os.Args[1] != "backup" {
		fmt.Println("Expected 'add','delete','backup','restore', 'sync' commands and its arguments. Run a command with -help for more info.")
		os.Exit(1)
	} else {
		src.ExecuteGoConsulKV()
	}

}
