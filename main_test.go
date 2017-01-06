package main_test

import (
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var (
	cliPath           string
	session           *Session
	err               error
	args              []string
	accessKeyId       string
	secretAccessKey   string
	region            string
	dateFormatRegex   string
	blockingProxyPath string
	openProxyPath     string
	proxyCommand      *exec.Cmd
	expectedExitCode  int
)

func setEnv() {
	os.Setenv("AWS_ACCESS_KEY_ID", accessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", secretAccessKey)
	region = "eu-west-1"
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

var _ = Describe("SchedLoad", func() {

	BeforeSuite(func() {

		accessKeyId = os.Getenv("TEST_AWS_ACCESS_KEY_ID")
		secretAccessKey = os.Getenv("TEST_AWS_SECRET_ACCESS_KEY")
		Ω(accessKeyId).ShouldNot(BeEmpty(), "You must set TEST_AWS_ACCESS_KEY_ID environment variable")
		Ω(secretAccessKey).ShouldNot(BeEmpty(), "You must set TEST_AWS_SECRET_ACCESS_KEY environment variable")
		cliPath, err = Build("github.com/dhrapson/sched-load")
		Ω(err).ShouldNot(HaveOccurred(), "Error building main source")
		blockingProxyPath, err = Build("github.com/dhrapson/sched-load/fixtures/blockingproxy")
		Ω(err).ShouldNot(HaveOccurred(), "Error building blockingproxy source")
		openProxyPath, err = Build("github.com/dhrapson/sched-load/fixtures/openproxy")
		Ω(err).ShouldNot(HaveOccurred(), "Error building openproxy source")

		SetDefaultEventuallyTimeout(30 * time.Second)
		dateFormatRegex = "[0-9]{4}/[0-9]{2}/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}"
	})

	AfterSuite(func() {
		CleanupBuildArtifacts()
	})

	Describe("invoking correctly", func() {

		BeforeEach(func() {
			setEnv()
			expectedExitCode = 0
		})

		AfterEach(func() {
			unsetEnv()
		})

		JustBeforeEach(func() {
			command := exec.Command(cliPath, args...)
			session, err = Start(command, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred(), "Error running CLI: "+cliPath)
			Eventually(session).Should(Exit(expectedExitCode), cliPath+" exited with non-zero error code")
		})

		Context("When run with status argument", func() {
			BeforeEach(func() {
				args = []string{"--region", region, "--integrator", "test-integrator", "--client", "test-client", "status"}
			})

			It("exits nicely", func() {
				Ω(session.Err).Should(Say(dateFormatRegex + " connected"))
			})
		})

		Context("When run with upload argument", func() {
			BeforeEach(func() {
				args = []string{"--region", region, "--integrator", "test-integrator", "--client", "test-client", "upload", "-f", "iaas/fixtures/test-file.csv"}
			})

			It("exits nicely", func() {
				Ω(session.Err).Should(Say(dateFormatRegex + " uploaded test-file.csv"))
			})
		})

		Context("When managing schedules", func() {
			Context("When setting daily", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", "test-integrator", "--client", "test-client", "set-schedule", "daily"}
				})

				It("indicates success", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " Set daily schedule: true"))
				})
			})

			Context("When showing daily", func() {
				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", "test-integrator", "--client", "test-client", "schedule"}
				})

				It("exits nicely", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " existing schedule: DAILY"))
				})
			})
		})

		Context("When run with help argument", func() {
			BeforeEach(func() {
				args = []string{"help"}
			})

			It("exits nicely", func() {
				Ω(session.Out).Should(Say("help"))
			})
		})

		Context("When run with an open proxy server", func() {

			BeforeEach(func() {
				args = []string{"--region", region, "--integrator", "test-integrator", "--client", "test-client", "status"}
				setEnv()
				// NB. use the openproxy port of 56565
				os.Setenv("HTTP_PROXY", "localhost:56565")
				proxyCommand = runProxyServer(openProxyPath)
			})

			AfterEach(func() {
				unsetEnv()
				killProxyServer(proxyCommand)
			})

			It("exits with zero error code", func() {
				Ω(session.Err).Should(Say(dateFormatRegex + " connected"))
			})

		})

		Context("When run with a blocking proxy server", func() {

			BeforeEach(func() {
				args = []string{"--region", region, "--integrator", "test-integrator", "--client", "test-client", "status"}
				setEnv()
				// NB. use the openproxy port of 56565
				os.Setenv("HTTP_PROXY", "localhost:56565")
				proxyCommand = runProxyServer(blockingProxyPath)
				expectedExitCode = 1
			})

			AfterEach(func() {
				unsetEnv()
				killProxyServer(proxyCommand)
			})

			It("exits with zero error code", func() {
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

				It("exits with non-zero error code", func() {
					Ω(session.Err).Should(Say(dateFormatRegex + " Invalid"))
				})
			})

			Context("When run with a missing proxy server", func() {

				BeforeEach(func() {
					args = []string{"--region", region, "--integrator", "test-integrator", "--client", "test-client", "status"}
					setEnv()
					// NB. Attempt to choose a port that is not otherwise in use
					os.Setenv("HTTP_PROXY", "localhost:45532")
				})

				AfterEach(func() {
					unsetEnv()
				})

				It("exits with non-zero error code", func() {
					Ω(session.Err).Should(Say("error connecting to proxy"))
				})

			})

			Context("When run without AWS creds", func() {

				BeforeEach(func() {
					args = []string{"status"}
				})

				It("exits with non-zero error code", func() {
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
