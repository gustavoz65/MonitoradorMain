package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gustavoz65/MoniMaster/internal/app"
)

var version = "3.0.0"

func main() {
	app.Version = version
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("MoniMaster CLI " + version)
		return
	}
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
