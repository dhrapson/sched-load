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
		var details iaas.IaaSAccountDetails

		JustBeforeEach(func() {
			details, err = controller.Status()
		})

		Context("when the IaaS is connecting", func() {
			BeforeEach(func() {
				iaasClient = IaaSClientMock{
					AccountDetail: iaas.IaaSAccountDetails{"AccountId": "123", "ClientId": "456", "IntegratorId": "789", "ConnectionType": "client"},
				}
			})
			It("gives status connected", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(details["AccountId"]).Should(Equal("123"))
				Ω(details["ClientId"]).Should(Equal("456"))
				Ω(details["IntegratorId"]).Should(Equal("789"))
				Ω(details["ConnectionType"]).Should(Equal("client"))
			})
		})

		Context("when the IaaS is not connecting", func() {
			BeforeEach(func() {
				iaasClient = IaaSClientMock{Err: errors.New("InvalidAccessKeyId")}
			})
			It("throws an error and returns the right status", func() {
				Ω(err).Should(HaveOccurred())
				Ω(details).Should(BeNil())
			})
		})
	})

	Describe("the CreateClientUser operation", func() {
		var result iaas.IaaSCredentials
		JustBeforeEach(func() {
			result, err = controller.CreateClientUser()
		})

		Context("when the IaaS is connecting", func() {

			BeforeEach(func() {
				iaasClient = IaaSClientMock{
					Credentials: iaas.AwsCredentials{
						AccessKeyId:     "abc",
						SecretAccessKey: "123",
					},
				}
			})

			It("indicates that the account was created", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result.String()).Should(Equal("AccessKeyId: abc, SecretAccessKey: 123"))
			})
		})

		Context("when the IaaS is not connecting", func() {
			BeforeEach(func() {
				iaasClient = IaaSClientMock{Err: errors.New("InvalidAccessKeyId")}
			})
			It("throws an error and returns the right result", func() {
				Ω(err).Should(HaveOccurred())
				Ω(result).Should(BeNil())
			})
		})
	})

	Describe("the DeleteClientUser operation", func() {
		var force bool
		var result bool

		JustBeforeEach(func() {
			force = false
			result, err = controller.DeleteClientUser(force)
		})

		Context("when deleting an existing account", func() {

			BeforeEach(func() {
				iaasClient = IaaSClientMock{Success: true}
			})

			It("indicates that the account was deleted", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(BeTrue())
			})
		})

		Context("when deleting a non-existant account", func() {

			BeforeEach(func() {
				iaasClient = IaaSClientMock{Success: false}
			})

			It("indicates that the account was deleted", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(BeFalse())
			})
		})

		Context("when deleting with force", func() {

			BeforeEach(func() {
				force = true
				iaasClient = IaaSClientMock{Success: true}
			})

			It("indicates that the account was deleted", func() {
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
			})
		})
	})

	Describe("the ImmediateDataFileCollectionStatus operation", func() {
		var result bool
		JustBeforeEach(func() {
			result, err = controller.ImmediateDataFileCollectionStatus()
		})

		Context("when the IaaS is connecting", func() {

			Context("when the setting is in place", func() {
				BeforeEach(func() {
					iaasClient = IaaSClientMock{Success: true}
				})
				It("indicates that it is enabled", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(BeTrue())
				})
			})

			Context("when the setting is NOT in place", func() {
				BeforeEach(func() {
					iaasClient = IaaSClientMock{Success: true}
				})
				It("indicates that is is disabled", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(BeTrue())
				})
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

	Describe("the EnableImmediateDataFileCollection operation", func() {
		var result bool
		JustBeforeEach(func() {
			result, err = controller.EnableImmediateDataFileCollection()
		})

		Context("when the IaaS is connecting", func() {

			Context("when the setting was previously in place", func() {
				BeforeEach(func() {
					iaasClient = IaaSClientMock{Success: false}
				})
				It("indicates that it was already enabled", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(BeFalse())
				})
			})

			Context("when the setting was NOT previously in place", func() {
				BeforeEach(func() {
					iaasClient = IaaSClientMock{Success: true}
				})
				It("indicates success in enabling it", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(BeTrue())
				})
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

	Describe("the DisableImmediateDataFileCollection operation", func() {
		var result bool
		JustBeforeEach(func() {
			result, err = controller.DisableImmediateDataFileCollection()
		})

		Context("when the IaaS is connecting", func() {

			Context("when the setting was previously in place", func() {
				BeforeEach(func() {
					iaasClient = IaaSClientMock{Success: true}
				})
				It("indicates success in disabling it", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(BeTrue())
				})
			})

			Context("when the setting was NOT previously in place", func() {
				BeforeEach(func() {
					iaasClient = IaaSClientMock{Success: true}
				})
				It("indicates that is was alread disabled", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(BeTrue())
				})
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

	Describe("the DeleteDataFile operation", func() {
		var result bool
		JustBeforeEach(func() {
			result, err = controller.DeleteDataFile("anyfile")
		})

		Context("when the IaaS is connecting", func() {

			Context("when the file was previously existing", func() {
				BeforeEach(func() {
					iaasClient = IaaSClientMock{Success: true}
				})
				It("indicates success in setting schedule", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(BeTrue())
				})
			})

			Context("when the file was NOT previously existing", func() {
				BeforeEach(func() {
					iaasClient = IaaSClientMock{Success: false}
				})
				It("indicates success in setting schedule", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(BeFalse())
				})
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

	Describe("the ListDataFiles operation", func() {
		var result []string
		JustBeforeEach(func() {
			result, err = controller.ListDataFiles()
		})

		Context("when the IaaS is connecting", func() {
			BeforeEach(func() {
				iaasClient = IaaSClientMock{FilesList: []string{"somefile", "INPUT/thefile", "INPUT/otherfile", "PROCESSED/anotherone"}}
			})
			It("finds the files within INPUT", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(Equal([]string{"INPUT/thefile", "INPUT/otherfile"}))
			})
		})

		Context("when the IaaS is not connecting", func() {
			BeforeEach(func() {
				iaasClient = IaaSClientMock{Err: errors.New("InvalidAccessKeyId")}
			})
			It("throws an error and returns the right result", func() {
				Ω(err).Should(HaveOccurred())
				Ω(result).Should(BeNil())
			})
		})
	})

	Describe("the UploadDataFile operation", func() {
		var result string
		JustBeforeEach(func() {
			result, err = controller.UploadDataFile("path/to/thefile")
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

	Describe("the RemoveSchedule operation", func() {
		var result bool
		JustBeforeEach(func() {
			result, err = controller.RemoveSchedule()
		})

		Context("when the IaaS is connecting", func() {

			Context("when the schedule was previously set", func() {
				BeforeEach(func() {
					iaasClient = IaaSClientMock{Success: true}
				})
				It("indicates success in setting schedule", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(BeTrue())
				})
			})

			Context("when the schedule was NOT previously set", func() {
				BeforeEach(func() {
					iaasClient = IaaSClientMock{Success: false}
				})
				It("indicates success in setting schedule", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(BeFalse())
				})
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

	Describe("the GetSchedule operation", func() {
		var result string
		JustBeforeEach(func() {
			result, err = controller.GetSchedule()
		})

		Context("when the IaaS is connecting", func() {

			Context("when no schedule is set", func() {
				BeforeEach(func() {
					iaasClient = IaaSClientMock{}
				})
				It("shows that none is set", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(Equal("NONE"))
				})
			})

			Context("when a daily schedule is set", func() {
				BeforeEach(func() {
					iaasClient = IaaSClientMock{FilesList: []string{"INPUT/somefile.txt.", "DAILY_SCHEDULE", "PROCESSED/someotherfile.txt"}}
				})
				It("returns a daily status", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(Equal("DAILY"))
				})
			})
		})

		Context("when the IaaS is not connecting", func() {
			BeforeEach(func() {
				iaasClient = IaaSClientMock{Err: errors.New("InvalidAccessKeyId")}
			})
			It("throws an error and returns the right result", func() {
				Ω(err).Should(HaveOccurred())
				Ω(result).Should(Equal("ERROR"))
			})
		})
	})
})
