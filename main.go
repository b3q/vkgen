package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/k0kubun/pp"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v2"
)

func generateSchemaCmd(c *cli.Context) error {
	return NewGenerator(
		c.Bool("nofmt"),
		c.Bool("nogoify"),
		c.Bool("debug"),
	).Generate()
}

func generateMethodCmd(c *cli.Context) error {
	methodName := c.Args().First()
	b, err := ioutil.ReadFile("methods.json")
	if err != nil {
		return err
	}

	m := gjson.ParseBytes(b).Get(`methods.#(name=="` + methodName + `")`)
	if !m.Exists() {
		return fmt.Errorf("invalid method name")
	}

	pp.Println(m)
	return nil
}

func generateObjectCmd(c *cli.Context) error {
	objectName := c.Args().First()
	b, err := ioutil.ReadFile("objects.json")
	if err != nil {
		return err
	}

	obj := gjson.ParseBytes(b).Get("definitions." + objectName)
	if !obj.Exists() {
		return fmt.Errorf("invalid object name")
	}

	pp.Println(obj)
	return nil
}

func generateResponseCmd(c *cli.Context) error {
	responseName := c.Args().First()
	b, err := ioutil.ReadFile("responses.json")
	if err != nil {
		return err
	}

	resp := gjson.ParseBytes(b).Get("definitions." + responseName)
	if !resp.Exists() {
		return fmt.Errorf("invalid response name")
	}

	pp.Println(resp)
	return nil
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
		Commands: []*cli.Command{
			// {
			// 	Name:  "schema",
			// 	Usage: "generate sources from VK schema",
			// 	Flags: []cli.Flag{
			// 		&cli.StringFlag{
			// 			Name: "path",
			// 			Usage: "sets the schema path (can point to schema file or folder which contains schema files)\n\t" +
			// 				"if no path is given, generates sources directly from master branch of https://github.com/VKCOM/vk-api-schema",
			// 		},
			// 		&cli.StringFlag{
			// 			Name:    "output",
			// 			Aliases: []string{"o", "out"},
			// 			Value:   "generated",
			// 			Usage:   "sets the output directory path for generated code",
			// 		},
			// 	},
			// 	Action: generateSchemaCmd,
			// },
			{
				Name:    "method",
				Aliases: []string{"m"},
				Usage:   "generate source for specified method",
				Action:  generateMethodCmd,
			},
			{
				Name:    "object",
				Aliases: []string{"o", "obj"},
				Usage:   "generate source for specified object",
				Action:  generateObjectCmd,
			},
			{
				Name:    "response",
				Aliases: []string{"r", "resp"},
				Usage:   "generate source for specified response",
				Action:  generateResponseCmd,
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
