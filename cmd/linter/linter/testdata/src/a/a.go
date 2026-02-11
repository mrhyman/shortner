package a

import (
	"log"
	"os"
)

func f() {
	panic("boom")      // want "panic\\(\\) usage is forbidden"
	log.Fatal("fatal") // want "log.Fatal/os.Exit is allowed only in main.main"
	os.Exit(1)         // want "log.Fatal/os.Exit is allowed only in main.main"
}
