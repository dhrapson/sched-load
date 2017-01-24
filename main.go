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
	var force bool

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

				iaasClient := iaas.AwsClient{Region: region, ClientId: clientId}
				ctrler := controller.Controller{Client: iaasClient}
				details, err := ctrler.Status()
				if err != nil {
					log.Fatalf("Error connecting: %s\n", err.Error())
				}
				log.Println("connected to IaaS")
				if details["ClientId"] == "" {
					log.Println("Connected as Integrator without a valid client")
				} else {
					log.Println("Client ID: " + details["ClientId"])
				}
				log.Println("Integrator ID: " + details["IntegratorId"])
				log.Println("Account ID: " + details["AccountId"])

				return nil
			},
		},
		{
			Name:    "client",
			Aliases: []string{"c"},
			Usage:   "manage client accounts",
			Subcommands: []cli.Command{
				{
					Name:    "delete",
					Aliases: []string{"d"},
					Usage:   "remove a client account, optionally also remove all the uploaded data files",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:        "force, f",
							Usage:       "force deletion of the user accounts uploaded files",
							Destination: &force,
						},
					},
					Action: func(c *cli.Context) error {

						iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
						controller := controller.Controller{Client: iaasClient}

						wasPreExisting, err := controller.DeleteClientUser(force)
						if err != nil {
							log.Fatalf("Error deleting client account, %s\n", err.Error())
						}
						if wasPreExisting {
							if force {
								log.Printf("deleted account %s, left data files in place\n", clientId)
							} else {
								log.Printf("deleted account %s & all data files\n", clientId)
							}
						} else {
							log.Printf("%s account did not exist\n", clientId)
							if force {
								log.Printf("removed any data files for account %s\n", clientId)
							}
						}
						return nil
					},
				},
				{
					Name:    "create",
					Aliases: []string{"add"},
					Usage:   "create a client account",
					Action: func(c *cli.Context) error {

						iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
						controller := controller.Controller{Client: iaasClient}

						creds, err := controller.CreateClientUser()
						if err != nil {
							log.Fatalf("Error creating client user %s, %s\n", clientId, err.Error())
						}

						log.Printf("created account %s\n", clientId)
						log.Printf("Credentials are %s\n", creds.String())

						return nil
					},
				},
			},
		},
		{
			Name:    "data-file",
			Aliases: []string{"df"},
			Usage:   "manage data files",
			Subcommands: []cli.Command{
				{
					Name:    "delete",
					Aliases: []string{"d"},
					Usage:   "remove a remote data file",
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

						wasPreExisting, err := controller.DeleteDataFile(filePath)
						if err != nil {
							log.Fatalf("Error delete file %s, %s\n", filePath, err.Error())
						}
						if wasPreExisting {
							log.Printf("deleted %s\n", filePath)
						} else {
							log.Printf("%s did not exist\n", filePath)
						}

						return nil
					},
				},
				{
					Name:    "list-uploaded",
					Aliases: []string{"lu"},
					Usage:   "list remote unprocessed data files",
					Action: func(c *cli.Context) error {

						iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
						controller := controller.Controller{Client: iaasClient}

						files, err := controller.ListDataFiles()
						if err != nil {
							log.Fatalf("Error listing data files, %s\n", err.Error())
						}

						var filesList string

						if len(files) == 0 {
							filesList = "\tnone found"
						}
						for _, filePath := range files {
							filesList += "\t" + filePath + "\n"
						}
						log.Printf("listing files:\n%s", filesList)

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
							Usage:       "path to the local file",
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
			},
		},
		{
			Name:    "immediate-collection",
			Aliases: []string{"ic"},
			Usage:   "enable/disable immediate collection of uploaded data files",
			Subcommands: []cli.Command{
				{
					Name:  "status",
					Usage: "show the immediate data file collection status",
					Action: func(c *cli.Context) error {

						iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
						controller := controller.Controller{Client: iaasClient}

						status, err := controller.ImmediateDataFileCollectionStatus()
						if err != nil {
							log.Fatalf("Error connecting: %s\n", err.Error())
						}
						if status {
							log.Println("Immediate collection status is enabled")
						} else {
							log.Println("Immediate collection status is disabled")
						}
						return nil
					},
				},
				{
					Name:  "enable",
					Usage: "enable immediate data file collection",
					Action: func(c *cli.Context) error {

						iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
						controller := controller.Controller{Client: iaasClient}

						wasNewlySet, err := controller.EnableImmediateDataFileCollection()
						if err != nil {
							log.Fatalf("Error connecting: %s\n", err.Error())
						}
						if wasNewlySet {
							log.Println("Enabled immediate collection")
						} else {
							log.Println("Immediate collection was already enabled")
						}
						return nil
					},
				},
				{
					Name:  "disable",
					Usage: "disable immediate data file collection",
					Action: func(c *cli.Context) error {

						iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
						controller := controller.Controller{Client: iaasClient}

						wasPreExisting, err := controller.DisableImmediateDataFileCollection()
						if err != nil {
							log.Fatalf("Error connecting: %s\n", err.Error())
						}
						if wasPreExisting {
							log.Println("Disabled immediate collection")
						} else {
							log.Println("Immediate collection was already disabled")
						}
						return nil
					},
				},
			},
		},
		{
			Name:    "schedule",
			Aliases: []string{"s"},
			Usage:   "set a schedule for collection of uploaded data files",
			Subcommands: []cli.Command{
				{
					Name:  "status",
					Usage: "show the schedule status",
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
					Name:  "daily",
					Usage: "set a daily schedule",
					Action: func(c *cli.Context) error {

						iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
						controller := controller.Controller{Client: iaasClient}

						wasPreExisting, err := controller.SetSchedule("DAILY")
						if err != nil {
							log.Fatalf("Error connecting: %s\n", err.Error())
						}
						if wasPreExisting {
							log.Println("Set daily schedule")
						} else {
							log.Println("Daily schedule was already set")
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

						wasPreExisting, err := controller.RemoveSchedule()
						if err != nil {
							log.Fatalf("Error connecting: %s\n", err.Error())
						}
						if wasPreExisting {
							log.Println("Removed schedule")
						} else {
							log.Println("No schedule existed to remove")
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
