package main

import (
	"log"
	"os"

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
			Name:    "data-file",
			Aliases: []string{"df"},
			Usage:   "manage data files",
			Subcommands: []cli.Command{
				{
					Name:  "delete",
					Usage: "remove a remote data file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "remote, r",
							Usage:       "remote file path",
							Destination: &filePath,
						},
					},
					Action: func(c *cli.Context) error {

						iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
						controller := controller.Controller{Client: iaasClient}

						if wasPreExisting, err := controller.DeleteDataFile(filePath); err != nil {
							log.Fatalf("Error delete file %s, %s\n", fileName, err.Error())
						} else {
							if wasPreExisting {
								log.Printf("deleted %s\n", fileName)
							} else {
								log.Printf("%s did not exist\n", fileName)
							}
						}

						return nil
					},
				}
			}
		},
		{
			Name:    "upload",
			Aliases: []string{"u"},
			Usage:   "upload a data file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "file, f",
					Usage:       "path to the file",
					Destination: &filePath,
				},
			},
			Action: func(c *cli.Context) error {

				iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
				controller := controller.Controller{Client: iaasClient}

				if fileName, err := controller.UploadDataFile(filePath); err != nil {
					log.Fatalf("Error uploading file %s, %s\n", fileName, err.Error())
				} else {
					log.Printf("uploaded %s\n", fileName)
				}

				return nil
			},
		},
		{
			Name:  "schedule",
			Usage: "show the schedule for collection of uploaded data files",
			Action: func(c *cli.Context) error {

				iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
				controller := controller.Controller{Client: iaasClient}

				if schedule, err := controller.GetSchedule(); err != nil {
					log.Fatalf("Error connecting: %s\n", err.Error())
				} else {
					log.Println("existing schedule: " + schedule)
				}
				return nil
			},
		},
		{
			Name:    "set-schedule",
			Aliases: []string{"ss"},
			Usage:   "set a schedule for collection of uploaded data files",
			Subcommands: []cli.Command{
				{
					Name:  "daily",
					Usage: "set a daily schedule",
					Action: func(c *cli.Context) error {

						iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
						controller := controller.Controller{Client: iaasClient}

						if wasPreExisting, err := controller.SetSchedule("DAILY"); err != nil {
							log.Fatalf("Error connecting: %s\n", err.Error())
						} else {

							if wasPreExisting {
								log.Println("Set daily schedule")
							} else {
								log.Println("Daily schedule was already set")
							}
						}
						return nil
					},
				},
				{
					Name:  "none",
					Usage: "remove schedule",
					Action: func(c *cli.Context) error {

						iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
						controller := controller.Controller{Client: iaasClient}

						if wasPreExisting, err := controller.RemoveSchedule(); err != nil {
							log.Fatalf("Error connecting: %s\n", err.Error())
						} else {
							if wasPreExisting {
								log.Println("Removed schedule")
							} else {
								log.Println("No schedule existed to remove")
							}
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
