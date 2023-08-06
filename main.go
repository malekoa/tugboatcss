package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "tugboatcss",
		Usage: "A CSS utility class generator",
		Action: func(ctx *cli.Context) error {
			if lexemes, err := LexFileAtPath("./sidebar.html"); err != nil {
				log.Fatal(err)
			} else {
				fmt.Println(lexemes)
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
