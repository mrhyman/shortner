package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("ok")
	os.Exit(1)
}

func helper() {
	log.Fatal("nope") // want "log.Fatal/os.Exit is allowed only in main.main"
}
