package main

import (
	"bufio"
	"fmt"
	. "github.com/SealNTibbers/GotalkInterpreter/evaluator"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func SetupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		os.Exit(0)
	}()
}

func main() {
	vm := NewSmalltalkWorkspace()
	consoleReader := bufio.NewReader(os.Stdin)
	fmt.Println("Hello seeker!")
	SetupCloseHandler()

	for {
		input, err := consoleReader.ReadString('\n') // this will prompt the user for input
		if err == nil {
			input := strings.Split(input, "\n")[0]
			result := vm.EvaluateToInterface(input)
			fmt.Print(">>> ")
			fmt.Println(result)
		}
	}
}
