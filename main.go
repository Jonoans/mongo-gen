package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jonoans/mongo-gen/codegen"
	"github.com/jonoans/mongo-gen/config"

	cli "github.com/urfave/cli/v2"
)

var generateCmd = &cli.Command{
	Name:  "generate",
	Usage: "Generate models",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "file",
			Aliases:     []string{"f"},
			Usage:       "Config file",
			DefaultText: "orm.yml",
		},
	},
	Action: func(c *cli.Context) error {
		configFilename := c.String("file")
		config := config.ParseConfig(configFilename)
		codegen.Generate(config)
		return nil
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "Go MongoORM"
	app.Usage = "Generate usable model code from struct models"
	app.Action = func(c *cli.Context) error {
		fmt.Println("Setup your orm.yml file and run `go run github.com/jonoans/mongo-gen generate` to begin!")
		return nil
	}
	app.Commands = []*cli.Command{
		generateCmd,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
