package main

import (
	"log"
	"os"

	generator "github.com/mrhyman/shortner/cmd/reset/reset"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if err := generator.Generate(root); err != nil {
		log.Fatal(err)
	}
}
