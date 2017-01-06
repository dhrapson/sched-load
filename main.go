package main

import (
	"log"
	"os"
	"strconv"

	"github.com/dhrapson/sched-load/controller"
	"github.com/dhrapson/sched-load/iaas"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	var region string
	var integratorId string
	var clientId string
	var filePath string

	app.Name = "sched-load"
	app.Usage = "uploads files to public IaaS & publishes a schedule for regular file uploads"

	flags := []cli.Flag{
		cli.StringFlag{
			Name:        "region, r",
			Usage:       "public IaaS region for storing the files",
			Destination: &region,
		},
		cli.StringFlag{
			Name:        "integrator, i",
			Usage:       "identifier for the integrator",
			Destination: &integratorId,
		},
		cli.StringFlag{
			Name:        "client, c",
			Usage:       "identifier for the client",
			Destination: &clientId,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "status",
			Aliases: []string{"s"},
			Usage:   "show status of connection and schedule",
			Action: func(c *cli.Context) error {

				iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
				controller := controller.Controller{Client: iaasClient}

				if status, err := controller.Status(); err != nil {
					log.Fatalf("Error connecting: %s\n", err.Error())
				} else {
					log.Println(status)
				}
				return nil
			},
		},
		{
			Name:    "upload",
			Aliases: []string{"u"},
			Usage:   "upload a data file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "file, f",
					Usage:       "identifier for the client",
					Destination: &filePath,
				},
			},
			Action: func(c *cli.Context) error {

				iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
				controller := controller.Controller{Client: iaasClient}

				if fileName, err := controller.UploadFile(filePath); err != nil {
					log.Fatalf("Error uploading file %s, %s\n", fileName, err.Error())
				} else {
					log.Printf("uploaded %s\n", fileName)
				}

				return nil
			},
		},
		{
			Name:  "schedule",
			Usage: "schedule the collection of uploaded data files",
			Subcommands: []cli.Command{
				{
					Name:  "daily",
					Usage: "set a daily schedule",
					Action: func(c *cli.Context) error {

						iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
						controller := controller.Controller{Client: iaasClient}

						if status, err := controller.SetSchedule("DAILY"); err != nil {
							log.Fatalf("Error connecting: %s\n", err.Error())
						} else {
							log.Println("Set daily schedule: " + strconv.FormatBool(status))
						}
						return nil
					},
				},
			},
		},
	}

	app.Flags = flags

	app.CommandNotFound = func(c *cli.Context, command string) {
		log.Printf("Invalid command '%s'\n\n", command)
		cli.ShowAppHelp(c)
		os.Exit(1)
	}
	app.Run(os.Args)
}
