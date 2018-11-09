package main

import (
	"bufio"
	"fmt"
	. "github.com/SealNTibbers/GotalkInterpreter/evaluator"
	"os"
	"strings"
)

func main() {
	vm := NewSmalltalkVM()
	consoleReader := bufio.NewReader(os.Stdin)
	fmt.Println("Hello seeker!")

	for true {
		input, err := consoleReader.ReadString('\n') // this will prompt the user for input
		if err == nil {
			input := strings.Split(input, "\n")[0]
			result := vm.EvaluateToInterface(input)
			fmt.Print(">>> ")
			fmt.Println(result)
		}
	}
}
