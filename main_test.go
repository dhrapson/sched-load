package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("SchedLoad", func() {

	var (
		cliPath string
		session *Session
		err     error
		args    []string
	)

	BeforeSuite(func() {
		cliPath, err = Build("github.com/dhrapson/sched-load")
		Ω(err).ShouldNot(HaveOccurred(), "Error building source")
	})

	AfterSuite(func() {
		CleanupBuildArtifacts()
	})

	Describe("invoking correctly", func() {

		JustBeforeEach(func() {
			command := exec.Command(cliPath, args...)
			session, err = Start(command, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred(), "Error running CLI: "+cliPath)
			Eventually(session).Should(Exit(0), cliPath+" exited with non-zero error code")
		})

		Context("When run with status argument", func() {
			BeforeEach(func() {
				args = []string{"status"}
			})

			It("exits nicely", func() {
				Ω(session.Out).Should(Say("status"))
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
			Eventually(session).Should(Exit(1), cliPath+" exited with unexpected error code")
		})

		Context("When run with an invalid argument", func() {
			BeforeEach(func() {
				args = []string{"foo"}
			})

			It("exits with non-zero error code", func() {
				Ω(session.Err).Should(Say("Invalid"))
			})

		})
	})
})
