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

var _ = Describe("SchedLoad", func() {

	var (
		cliPath         string
		session         *Session
		err             error
		args            []string
		accessKeyId     string
		secretAccessKey string
		dateFormatRegex string
	)

	BeforeSuite(func() {

		accessKeyId = os.Getenv("TEST_AWS_ACCESS_KEY_ID")
		secretAccessKey = os.Getenv("TEST_AWS_SECRET_ACCESS_KEY")
		Ω(accessKeyId).ShouldNot(BeEmpty(), "You must set TEST_AWS_ACCESS_KEY_ID environment variable")
		Ω(secretAccessKey).ShouldNot(BeEmpty(), "You must set TEST_AWS_SECRET_ACCESS_KEY environment variable")
		cliPath, err = Build("github.com/dhrapson/sched-load")
		Ω(err).ShouldNot(HaveOccurred(), "Error building source")

		SetDefaultEventuallyTimeout(30 * time.Second)
		dateFormatRegex = "[0-9]{4}/[0-9]{2}/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}"
	})

	AfterSuite(func() {
		CleanupBuildArtifacts()
	})

	Describe("invoking correctly", func() {

		BeforeEach(func() {
			os.Setenv("AWS_ACCESS_KEY_ID", accessKeyId)
			os.Setenv("AWS_SECRET_ACCESS_KEY", secretAccessKey)
		})

		AfterEach(func() {
			os.Unsetenv("AWS_ACCESS_KEY_ID")
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		})

		JustBeforeEach(func() {
			command := exec.Command(cliPath, args...)
			session, err = Start(command, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred(), "Error running CLI: "+cliPath)
			Eventually(session).Should(Exit(0), cliPath+" exited with non-zero error code")
		})

		Context("When run with status argument", func() {
			BeforeEach(func() {
				args = []string{"--region", "eu-west-1", "--integrator", "test-integrator", "--client", "test-client", "status"}
			})

			It("exits nicely", func() {
				Ω(session.Err).Should(Say(dateFormatRegex+" connected"))
			})
		})

		Context("When run with upload argument", func() {
			BeforeEach(func() {
				args = []string{"--region", "eu-west-1", "--integrator", "test-integrator", "--client", "test-client", "upload", "-f", "iaas/fixtures/test-file.csv"}
			})

			It("exits nicely", func() {
				Ω(session.Err).Should(Say(dateFormatRegex+" uploaded test-file.csv"))
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
	})

	Describe("invoking incorrectly", func() {
		JustBeforeEach(func() {
			command := exec.Command(cliPath, args...)
			session, err = Start(command, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred(), "Error running CLI: "+cliPath)
		})

		Context("When run using commands", func() {

			Context("When run with an invalid command", func() {

				BeforeEach(func() {
					args = []string{"foo"}
				})

				It("exits with non-zero error code", func() {
					Eventually(session).Should(Exit(1), cliPath+" exited with unexpected error code")
					Ω(session.Err).Should(Say(dateFormatRegex+" Invalid"))
				})
			})

			Context("When run without AWS creds", func() {

				BeforeEach(func() {
					args = []string{"status"}
				})

				It("exits with non-zero error code", func() {
					Eventually(session).Should(Exit(1), cliPath+" exited with unexpected error code")
					Ω(session.Err).Should(Say(dateFormatRegex+" Credentials not set"))
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
