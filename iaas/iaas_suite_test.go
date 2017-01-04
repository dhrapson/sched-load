package iaas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestIaas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Iaas Suite")
}
