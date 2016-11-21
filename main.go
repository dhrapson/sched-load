package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/urfave/cli"
)

func isValidActionCommand(actionCommand string) bool {
	actions := []string{"status", "help"}
	sort.Strings(actions)
	i := sort.SearchStrings(actions, actionCommand)
	if i < len(actions) && actions[i] == actionCommand {
		return true
	}
	return false
}

func Run(args []string) {
	fmt.Println("Running cli with", args)
	app := cli.NewApp()
	app.Name = "sched-load"
	app.Usage = "uploads files to public IaaS & publishes a schedule for regular file uploads"

	app.Action = func(c *cli.Context) error {
		actionName := "help"
		if c.NArg() > 0 {
			actionName = c.Args().Get(0)
		}
		if isValidActionCommand(actionName) {
			fmt.Println("Running ", actionName)
		} else {
			log.Fatal("Invalid action command: ", actionName)
		}
		return nil
	}

	app.Run(args)
}

func main() {
	Run(os.Args)
}
