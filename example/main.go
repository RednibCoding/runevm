package main

import (
	"fmt"
	"os"

	"github.com/RednibCoding/runevm"
)

func main() {

	args := []string{"example.exe", "test.rune"}
	// args := os.Args
	if len(args) < 2 {
		fmt.Println("USAGE: rune <sourcefile>")
		os.Exit(0)
	}
	source, err := os.ReadFile(args[1])
	if err != nil {
		fmt.Printf("ERROR: Can't find source file '%s'.\n", args[1])
		os.Exit(0)
	}

	filepath := args[1]

	vm := runevm.NewRuneVM()
	vm.Run(string(source), filepath)
}
