package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "sched-load"
	app.Usage = "uploads files to public IaaS & publishes a schedule for regular file uploads"

	app.Commands = []cli.Command{
		{
			Name:    "status",
			Aliases: []string{"s"},
			Usage:   "show status of connection and schedule",
			Action: func(c *cli.Context) error {
				fmt.Println("status")
				return nil
			},
		},
	}

	app.CommandNotFound = func(c *cli.Context, command string) {
		log.Printf("Invalid command '%s'\n\n", command)
		cli.ShowAppHelp(c)
		os.Exit(1)
	}
	app.Run(os.Args)
}
