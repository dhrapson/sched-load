package main

import (
	"log"
	"os"

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
			Name:        "region",
			Usage:       "public IaaS region for storing the files",
			Destination: &region,
		},
		cli.StringFlag{
			Name:        "integrator",
			Usage:       "identifier for the integrator",
			Destination: &integratorId,
		},
		cli.StringFlag{
			Name:        "client",
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

				iaasClient := iaas.IaaSClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
				_, err := iaasClient.ListFiles()
				if err != nil {
					log.Fatalf("error: %v", err)
				}
				log.Println("connected")
				return nil
			},
		},
		{
			Name:    "upload",
			Aliases: []string{"u"},
			Usage:   "upload a data file",
			 Flags: []cli.Flag{
      	cli.StringFlag{
      		Name: "file, f",
      		Usage:       "identifier for the client",
					Destination: &filePath,
      	},

      },
			Action: func(c *cli.Context) error {

				iaasClient := iaas.IaaSClient{Region: region, IntegratorId: integratorId, ClientId: clientId}
				fileName, err := iaasClient.UploadFile(filePath); if err != nil {
					log.Fatalf("error: %v", err)
				}

				fileNames, err := iaasClient.ListFiles(); if err != nil {
					log.Fatalf("error: %v", err)
				}

				if (arrayContains(fileNames, fileName)) {
					log.Printf("uploaded %s\n", fileName)
				} else {
					log.Fatalf("unable to find uploaded file %s\n", fileName)
				}

				return nil
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

func arrayContains(haystack []string, needle string) bool {
	for _, hay := range haystack {
		if needle == hay {
			return true
		}
	}
	return false
}
