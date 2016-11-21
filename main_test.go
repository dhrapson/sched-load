package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func runCli(args ...string) (string, error) {
	cmd := exec.Command("sched-load", args...)
	outBytes, err := cmd.CombinedOutput()
	return string(outBytes), err
}

var _ = Describe("SchedLoad", func() {

	Describe("running the CLI", func() {
		Context("When run with status argument", func() {
			It("exits nicely", func() {
				stringArgs := []string{"status"}
				output, err := runCli(stringArgs...)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(ContainSubstring("status"))
			})

		})

		Context("When run with help argument", func() {
			It("exits nicely", func() {
				stringArgs := []string{"help"}
				output, err := runCli(stringArgs...)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(ContainSubstring("help"))
			})

		})

		Context("When run with an invalid argument", func() {
			It("exits with non-zero error code", func() {
				stringArgs := []string{"blah"}
				output, err := runCli(stringArgs...)
				Expect(err).To(HaveOccurred())
				Expect(output).To(ContainSubstring("Invalid"))
			})

		})
	})

})
