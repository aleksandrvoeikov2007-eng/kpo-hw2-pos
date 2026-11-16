package main

import (
	"fmt"
	"os"
	"pos/internal/adapters/console"
	"pos/internal/di"
)

func main() {
	container := di.Default()
	if err := console.New(container.Facade()).Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка чтения:", err)
		os.Exit(1)
	}
}
