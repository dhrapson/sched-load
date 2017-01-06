package controller_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/dhrapson/sched-load/controller"
	"github.com/dhrapson/sched-load/iaas"

	"errors"
)

var (
	controller Controller
	iaasClient iaas.IaaSClient
	err error
)
var _ = Describe("The controller", func() {

	JustBeforeEach(func() {
		controller = Controller{Client: iaasClient}
	})
	Describe("the Status operation", func() {
			var status string

		JustBeforeEach(func() {
			status, err = controller.Status()
		 })

		Context("when the IaaS is connecting", func() {
			BeforeEach(func() {
				iaasClient = IaaSClientMock{FilesList: []string{""}}
			})
			It("gives status connected", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(status).Should(Equal("connected"))
			})
		})
		Context("when the IaaS is not connecting", func() {
			BeforeEach(func() {
				iaasClient = IaaSClientMock{Err: errors.New("InvalidAccessKeyId")}
			})
			It("throws an error and returns the right status", func() {
				Ω(err).Should(HaveOccurred())
				Ω(status).Should(Equal("error"))
			})
		})
	})
})

