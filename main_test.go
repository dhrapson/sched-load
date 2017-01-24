package main_test

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/dhrapson/sched-load/controller"
	"github.com/dhrapson/sched-load/iaas"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	uuid "github.com/satori/go.uuid"
)

var (
	cliPath                   string
	session                   *Session
	err                       error
	args                      []string
	integratorAccessKeyId     string
	integratorSecretAccessKey string
	region                    string
	dateFormatRegex           string
	blockingProxyPath         string
	openProxyPath             string
	proxyCommand              *exec.Cmd
	expectedExitCode          int
	integratorName            string
	clientName                string
	uniqueId                  string
	clientCreds               iaas.IaaSCredentials
	ctrler                    controller.Controller
)

func setIntegratorEnv() {
	os.Setenv("AWS_ACCESS_KEY_ID", integratorAccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", integratorSecretAccessKey)
}

func setClientEnv() {
	os.Setenv("AWS_ACCESS_KEY_ID", clientCreds.Map()["AccessKeyId"])
	os.Setenv("AWS_SECRET_ACCESS_KEY", clientCreds.Map()["SecretAccessKey"])
}

func unsetEnv() {
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
}

func runProxyServer(path string) *exec.Cmd {
	command := exec.Command(path, args...)
	session, err = Start(command, GinkgoWriter, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred(), "Error running CLI: "+path)
	return command
}

func killProxyServer(cmd *exec.Cmd) {
	e := cmd.Process.Kill()
	Ω(e).ShouldNot(HaveOccurred(), "Error killing process: "+cmd.Path)
}

func runCommand(path string, exitCode int, argsList ...string) (*Session, error) {
	command := exec.Command(path, argsList...)
	sess, localError := Start(command, GinkgoWriter, GinkgoWriter)
	Ω(localError).ShouldNot(HaveOccurred(), "Error running CLI: "+cliPath)
	Eventually(sess).Should(Exit(exitCode), cliPath+" exited with non-zero error code")
	return sess, localError
}

func waitForAws() {
	// AWS takes time to store settings
	time.Sleep(15 * time.Second)
}

var _ = Describe("SchedLoad", func() {

	BeforeSuite(func() {

		region = "eu-west-1"
		integratorAccessKeyId = os.Getenv("TEST_AWS_ACCESS_KEY_ID")
		integratorSecretAccessKey = os.Getenv("TEST_AWS_SECRET_ACCESS_KEY")
		Ω(integratorAccessKeyId).ShouldNot(BeEmpty(), "You must set TEST_AWS_ACCESS_KEY_ID environment variable")
		Ω(integratorSecretAccessKey).ShouldNot(BeEmpty(), "You must set TEST_AWS_SECRET_ACCESS_KEY environment variable")
		cliPath, err = Build("github.com/dhrapson/sched-load")
		Ω(err).ShouldNot(HaveOccurred(), "Error building main source")
		blockingProxyPath, err = Build("github.com/dhrapson/sched-load/fixtures/blockingproxy")
		Ω(err).ShouldNot(HaveOccurred(), "Error building blockingproxy source")
		openProxyPath, err = Build("github.com/dhrapson/sched-load/fixtures/openproxy")
		Ω(err).ShouldNot(HaveOccurred(), "Error building openproxy source")

		SetDefaultEventuallyTimeout(30 * time.Second)
		dateFormatRegex = "[0-9]{4}/[0-9]{2}/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}"

		if os.Getenv("INTEGRATOR") != "" {
			integratorName = os.Getenv("INTEGRATOR")
		} else {
			integratorName = "myintegrator"
		}

		uniqueId = uuid.NewV4().String()
		clientName = uuid.NewV4().String()

		iaasClient := iaas.AwsClient{Region: region, IntegratorId: integratorName, ClientId: clientName}
		ctrler = controller.Controller{Client: iaasClient}
		setIntegratorEnv()
		clientCreds, err = ctrler.CreateClientUser()
		waitForAws()
		Ω(err).ShouldNot(HaveOccurred())
		unsetEnv()
	})

	AfterSuite(func() {
		CleanupBuildArtifacts()
		setIntegratorEnv()
		_, err = ctrler.DeleteClientUser(true)
		Ω(err).ShouldNot(HaveOccurred())
		unsetEnv()
	})

	JustBeforeEach(func() {
		log.Println("running", cliPath, args, "expecting", expectedExitCode)
		session, err = runCommand(cliPath, expectedExitCode, args...)
	})

	Describe("invoking integrator operations", func() {
		BeforeEach(func() {
			setIntegratorEnv()
			expectedExitCode = 0
		})

		AfterEach(func() {
			unsetEnv()
		})

		Context("When managing client accounts", func() {
			Context("When creating", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", uniqueId, "client", "create"}
				})

				It("says the right thing and exits nicely", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " created account " + uniqueId))
					Ω(session.Err).Should(Say(dateFormatRegex + " Credentials are"))
				})
			})

			Context("When deleting", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", uniqueId, "client", "delete"}
				})

				It("says the right thing and exits nicely", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " deleted account " + uniqueId))
				})
			})

			Context("When deleting forcefully", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", uniqueId, "client", "delete", "-f"}
				})

				It("says the right thing and exits nicely", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " " + uniqueId + " account did not exist"))
					Ω(session.Err).Should(Say(dateFormatRegex + " removed any data files for account " + uniqueId))
				})
			})
		})

	})
	Describe("invoking client operations", func() {

		BeforeEach(func() {
			setClientEnv()
			expectedExitCode = 0
		})

		AfterEach(func() {
			unsetEnv()
		})

		Context("When run with status argument", func() {
			BeforeEach(func() {
				args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "status"}
			})

			It("exits nicely", func() {
				Ω(session.Err).Should(Say(dateFormatRegex + " connected"))
				Ω(session.Err).Should(Say(dateFormatRegex + " Client ID: " + clientName))
				Ω(session.Err).Should(Say(dateFormatRegex + " Integrator ID: " + integratorName))
				Ω(session.Err).Should(Say(dateFormatRegex + " Account ID: " + accountId))
			})
		})

		Context("When managing data files", func() {
			Context("When uploading", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "df", "upload", "-f", "iaas/fixtures/test-file.csv"}
				})

				It("exits nicely", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " uploaded INPUT/test-file.csv"))
				})
			})

			Context("When listing", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "df", "list-uploaded"}
				})

				It("finds the uploaded file", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + ` listing files:
	INPUT/test-file.csv`))
				})
			})

			Context("When deleting an existing file", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "data-file", "delete", "-r", "test-file.csv"}
				})

				It("exits nicely", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " deleted test-file.csv"))
				})
			})

			Context("When listing", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "df", "lu"}
				})

				It("finds nothing", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + ` listing files:
	none found`))
				})
			})

			Context("When deleting a non-existant file", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "data-file", "delete", "-r", "test-file.csv"}
				})

				It("exits nicely", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " test-file.csv did not exist"))
				})
			})
		})

		Context("When managing immediate file collection", func() {

			Context("When enabling immediate collection", func() {
				BeforeEach(func() {
					setupArgs := []string{"--region", region, "--integrator", integratorName, "--client", clientName, "immediate-collection", "disable"}
					runCommand(cliPath, 0, setupArgs...)
					waitForAws()
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "immediate-collection", "enable"}
				})

				It("enables immediate collection", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " Enabled immediate collection"))
				})
			})

			Context("When immediate collection is enabled", func() {
				BeforeEach(func() {
					enableArgs := []string{"--region", region, "--integrator", integratorName, "--client", clientName, "immediate-collection", "enable"}
					runCommand(cliPath, 0, enableArgs...)
					waitForAws()
				})

				Context("status command", func() {
					BeforeEach(func() {
						args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "immediate-collection", "status"}
					})
					It("shows status of enabled", func() {
						Ω(session.Err).Should(Say(dateFormatRegex + " Immediate collection status is enabled"))
					})
				})

				Context("enable command", func() {
					BeforeEach(func() {
						args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "immediate-collection", "enable"}
					})
					It("indicates that nothing was done", func() {
						Ω(session.Err).Should(Say(dateFormatRegex + " Immediate collection was already enabled"))
					})
				})
			})

			Context("When immediate collection is disabled", func() {
				BeforeEach(func() {
					enableArgs := []string{"--region", region, "--integrator", integratorName, "--client", clientName, "immediate-collection", "disable"}
					runCommand(cliPath, 0, enableArgs...)
					waitForAws()
				})

				Context("status command", func() {
					BeforeEach(func() {
						args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "immediate-collection", "status"}
					})
					It("shows status of enabled", func() {
						Ω(session.Err).Should(Say(dateFormatRegex + " Immediate collection status is disabled"))
					})
				})

				Context("enable command", func() {
					BeforeEach(func() {
						args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "immediate-collection", "disable"}
					})
					It("indicates that nothing was done", func() {
						Ω(session.Err).Should(Say(dateFormatRegex + " Immediate collection was already disabled"))
					})
				})
			})

			Context("When disabling immediate collection", func() {
				BeforeEach(func() {
					setupArgs := []string{"--region", region, "--integrator", integratorName, "--client", clientName, "immediate-collection", "enable"}
					runCommand(cliPath, 0, setupArgs...)
					waitForAws()
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "immediate-collection", "disable"}
				})

				It("disables immediate collection", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " Disabled immediate collection"))
				})
			})
		})

		Context("When managing schedules", func() {
			Context("When setting daily", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "schedule", "daily"}
				})

				It("indicates success", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " Set daily schedule"))
				})
			})

			Context("When showing existing schedule", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "schedule", "status"}
				})

				It("exits nicely", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " existing schedule: DAILY"))
				})
			})

			Context("When removing schedule", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "schedule", "none"}
				})

				It("indicates success", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " Removed schedule"))
				})
			})

			Context("When showing non-existing schedule", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "schedule", "status"}
				})

				It("exits nicely", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " existing schedule: NONE"))
				})
			})
		})

		Context("When run with help argument", func() {
			BeforeEach(func() {
				args = []string{"help"}
			})

			It("prints a nice help message", func() {
				Ω(session.Out).Should(Say("help"))
			})
		})

		Context("When run with an open proxy server", func() {

			BeforeEach(func() {
				args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "status"}
				setClientEnv()
				// NB. use the openproxy port of 56565
				os.Setenv("HTTP_PROXY", "localhost:56565")
				proxyCommand = runProxyServer(openProxyPath)
			})

			AfterEach(func() {
				unsetEnv()
				killProxyServer(proxyCommand)
			})

			It("is able to connect through the proxy server", func() {
				Ω(session.Err).Should(Say(dateFormatRegex + " connected"))
			})

		})

		Context("When run with a blocking proxy server", func() {

			BeforeEach(func() {
				args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "status"}
				setClientEnv()
				// NB. use the openproxy port of 56565
				os.Setenv("HTTP_PROXY", "localhost:56565")
				proxyCommand = runProxyServer(blockingProxyPath)
				expectedExitCode = 1
			})

			AfterEach(func() {
				unsetEnv()
				killProxyServer(proxyCommand)
			})

			It("throws an error", func() {
				Ω(session.Err).Should(Say(dateFormatRegex + " Error connecting: RequestError: send request failed"))
			})

		})
	})

	Describe("invoking incorrectly", func() {

		Context("When run using commands", func() {

			JustBeforeEach(func() {
				command := exec.Command(cliPath, args...)
				session, err = Start(command, GinkgoWriter, GinkgoWriter)
				Ω(err).ShouldNot(HaveOccurred(), "Error running CLI: "+cliPath)
				Eventually(session).Should(Exit(1), cliPath+" exited with unexpected error code")

			})

			Context("When run with an invalid command", func() {

				BeforeEach(func() {
					args = []string{"foo"}
				})

				It("indicates that the command was invalid", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " Invalid"))
				})
			})

			Context("When run with a missing proxy server", func() {

				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "status"}
					setClientEnv()
					// NB. Attempt to choose a port that is not otherwise in use
					os.Setenv("HTTP_PROXY", "localhost:45532")
				})

				AfterEach(func() {
					unsetEnv()
					os.Unsetenv("HTTP_PROXY")
				})

				It("throws an error", func() {
					Ω(session.Err).Should(Say("error connecting to proxy"))
				})

			})

			Context("When run without AWS creds", func() {

				BeforeEach(func() {
					unsetEnv()
					args = []string{"--region", region, "--integrator", integratorName, "--client", clientName, "status"}
				})

				It("indicates a credentials issue", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " Credentials not set"))
				})

			})
		})

		Context("When run with no arguments", func() {

			It("exits with non-zero error code", func() {
				command := exec.Command(cliPath)
				session, err = Start(command, GinkgoWriter, GinkgoWriter)
				Ω(err).ShouldNot(HaveOccurred(), "Error running CLI: "+cliPath)
				Eventually(session).Should(Exit(0), cliPath+" exited with unexpected error code")
				Ω(session.Out).Should(Say("NAME:"))
			})
		})

	})

})
