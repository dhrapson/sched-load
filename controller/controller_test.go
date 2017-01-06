package controller_test

import (
	. "github.com/dhrapson/sched-load/controller"
	"github.com/dhrapson/sched-load/iaas"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
)

var (
	controller Controller
	iaasClient iaas.IaaSClient
	err        error
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

	Describe("the Upload operation", func() {
		var result string
		JustBeforeEach(func() {
			result, err = controller.UploadFile("path/to/thefile")
		})

		Context("when the IaaS is connecting", func() {
			BeforeEach(func() {
				iaasClient = IaaSClientMock{FilesList: []string{"somefile", "thefile", "otherfile"}, FileName: "thefile"}
			})
			It("gives uploaded result", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(Equal("thefile"))
			})
		})

		Context("when the IaaS is not connecting", func() {
			BeforeEach(func() {
				iaasClient = IaaSClientMock{Err: errors.New("InvalidAccessKeyId")}
			})
			It("throws an error and returns the right result", func() {
				Ω(err).Should(HaveOccurred())
				Ω(result).Should(Equal("error"))
			})
		})
	})

	Describe("the SetSchedule operation", func() {
		var result bool
		JustBeforeEach(func() {
			result, err = controller.SetSchedule("DAILY")
		})

		Context("when the IaaS is connecting", func() {
			BeforeEach(func() {
				iaasClient = IaaSClientMock{FileName: "DAILY_SCHEDULE"}
			})
			It("gives set", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(BeTrue())
			})
		})

		Context("when the IaaS is not connecting", func() {
			BeforeEach(func() {
				iaasClient = IaaSClientMock{Err: errors.New("InvalidAccessKeyId")}
			})
			It("throws an error and returns the right result", func() {
				Ω(err).Should(HaveOccurred())
				Ω(result).Should(BeFalse())
			})
		})
	})
})
