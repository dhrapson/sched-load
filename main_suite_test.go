package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSchedLoad(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SchedLoad Main Suite")
}
