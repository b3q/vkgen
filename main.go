package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func generateSchemaCmd(c *cli.Context) error {
	objschema, err := ioutil.ReadFile("objects.json")
	if err != nil {
		return err
	}
	return NewGenerator(
		c.Bool("nofmt"),
		c.Bool("nogoify"),
		c.Bool("debug"),
		objschema,
	).Generate()
}

func main() {
	app := &cli.App{
		Name:  "vkgen",
		Usage: "generates Golang sources from VK Schema",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "nofmt",
				Usage: "disable code formatting",
			},
			&cli.BoolFlag{
				Name:  "nogoify",
				Usage: "disable names gopherization",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "print debug information",
			},
		},
		HideHelpCommand: true,
		Action:          generateSchemaCmd,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
