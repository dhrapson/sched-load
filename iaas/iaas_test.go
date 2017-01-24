package iaas_test

import (
	"io/ioutil"
	"os"
	"time"

	. "github.com/dhrapson/sched-load/iaas"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
)

var (
	client                    IaaSClient
	integratorAccessKeyId     string
	integratorSecretAccessKey string
	region                    string
	uniqueId                  string
	integratorName            string
	accountId                 string
	clientName                string
	clientCreds               IaaSCredentials
	err                       error
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

func waitForAws() {
	// AWS takes time to store settings
	time.Sleep(15 * time.Second)
}

var _ = Describe("The IaaS Client", func() {

	BeforeSuite(func() {
		integratorAccessKeyId = os.Getenv("TEST_AWS_ACCESS_KEY_ID")
		integratorSecretAccessKey = os.Getenv("TEST_AWS_SECRET_ACCESS_KEY")
		Ω(integratorAccessKeyId).ShouldNot(BeEmpty(), "You must set TEST_AWS_ACCESS_KEY_ID environment variable")
		Ω(integratorSecretAccessKey).ShouldNot(BeEmpty(), "You must set TEST_AWS_SECRET_ACCESS_KEY environment variable")
		if os.Getenv("INTEGRATOR") != "" {
			integratorName = os.Getenv("INTEGRATOR")
		} else {
			integratorName = "myintegrator"
		}

		if os.Getenv("ACCOUNT_ID") != "" {
			accountId = os.Getenv("ACCOUNT_ID")
		} else {
			accountId = "609701658665"
		}

		region = "eu-west-1"

		uniqueId = uuid.NewV4().String()
		clientName = uuid.NewV4().String()

		client := AwsClient{Region: region, ClientId: clientName}
		setIntegratorEnv()
		clientCreds, err = client.CreateClientUser()
		waitForAws()
		Ω(err).ShouldNot(HaveOccurred())
		unsetEnv()
	})

	AfterSuite(func() {
		setIntegratorEnv()
		client = AwsClient{Region: region, ClientId: clientName}
		_, err = client.DeleteClientUser(true)
		Ω(err).ShouldNot(HaveOccurred())
		unsetEnv()
	})

	Describe("Interacting with AWS", func() {

		var (
			result         []string
			localFilePath  string
			remoteFilePath string
			status         bool
		)

		Describe("the integrator-level operations", func() {

			BeforeEach(func() {
				setIntegratorEnv()
			})

			AfterEach(func() {
				unsetEnv()
			})

			JustBeforeEach(func() {
				client = AwsClient{ClientId: uniqueId, Region: region}
			})

			Context("when getting account status", func() {
				It("connects correctly & creates the user", func() {
					details, err := client.AccountDetails()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(details["AccountId"]).Should(Equal(accountId))
					Ω(details["IntegratorId"]).Should(Equal(integratorName))
				})
			})

			Context("when managing users", func() {

				It("connects correctly & creates the user", func() {
					credentials, err := client.CreateClientUser()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(credentials.String()).ShouldNot(BeNil())
				})

				It("deletes the user leaving files in place", func() {
					client.UploadFile("fixtures/test-file.csv", "test-file.csv")
					wasPreExisting, err := client.DeleteClientUser(false)
					Ω(err).ShouldNot(HaveOccurred())
					remainingFiles, err := client.ListFiles()
					Ω(err).ShouldNot(HaveOccurred())
					found := false
					for _, file := range remainingFiles {
						if file == "test-file.csv" {
							found = true
							break
						}
					}
					Ω(wasPreExisting).Should(BeTrue())
					Ω(found).Should(BeTrue())
				})

				It("deletes the user and all the files", func() {
					wasPreExisting, err := client.DeleteClientUser(true)
					Ω(err).ShouldNot(HaveOccurred())
					remainingFiles, err := client.ListFiles()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(len(remainingFiles)).Should(BeZero())
					Ω(wasPreExisting).Should(BeFalse())
				})
			})
		})

		Describe("the client-level operations", func() {

			JustBeforeEach(func() {
				client = AwsClient{ClientId: clientName, Region: region}
			})

			Describe("managing files", func() {
				var (
					tempDir string
				)

				Context("managing bucket notifications", func() {
					BeforeEach(func() {
						setClientEnv()
					})

					AfterEach(func() {
						unsetEnv()
					})

					It("connects correctly & adds the notification", func() {
						_, err = client.RemoveFileUploadNotification()
						Ω(err).ShouldNot(HaveOccurred())

						status, err = client.AddFileUploadNotification()
						Ω(err).ShouldNot(HaveOccurred())
						Ω(status).Should(BeTrue())
					})

					It("finds the added notification", func() {
						status, err = client.FileUploadNotification()
						Ω(err).ShouldNot(HaveOccurred())
						Ω(status).Should(BeTrue())

						status, err = client.AddFileUploadNotification()
						Ω(err).ShouldNot(HaveOccurred())
						Ω(status).Should(BeFalse())
					})

					It("removes the added notification", func() {
						status, err = client.RemoveFileUploadNotification()
						Ω(err).ShouldNot(HaveOccurred())
						Ω(status).Should(BeTrue())
					})

					It("does not find the removed notification", func() {
						status, err = client.FileUploadNotification()
						Ω(err).ShouldNot(HaveOccurred())
						Ω(status).Should(BeFalse())

						status, err = client.RemoveFileUploadNotification()
						Ω(err).ShouldNot(HaveOccurred())
						Ω(status).Should(BeFalse())
					})
				})

				Context("managing files with valid connection details", func() {
					BeforeEach(func() {
						setClientEnv()
						tempDir, err = ioutil.TempDir("", "iaas-uploading-files")
						Ω(err).ShouldNot(HaveOccurred())
					})

					AfterEach(func() {
						unsetEnv()
					})

					It("connects correctly & uploads the file", func() {
						remoteFilePath, err = client.UploadFile("fixtures/test-file.csv", "someother-file.csv")
						Ω(err).ShouldNot(HaveOccurred())
						Ω(remoteFilePath).Should(Equal("someother-file.csv"))

					})

					It("finds the file it just uploaded", func() {
						result, err = client.ListFiles()
						Ω(err).ShouldNot(HaveOccurred())
						Ω(result).Should(ContainElement("someother-file.csv"))
					})

					It("downloads the file and the contents match", func() {
						localFilePath, err := client.GetFile(remoteFilePath, tempDir)
						Ω(err).ShouldNot(HaveOccurred())
						contents, err := ioutil.ReadFile(localFilePath)
						Ω(err).ShouldNot(HaveOccurred())
						expectedContents, err := ioutil.ReadFile("fixtures/test-file.csv")
						Ω(err).ShouldNot(HaveOccurred())
						Ω(contents).Should(Equal(expectedContents))
					})

					It("deletes the file it just uploaded", func() {
						status, err = client.DeleteFile("someother-file.csv")
						Ω(err).ShouldNot(HaveOccurred())
						Ω(status).Should(BeTrue())
					})

					It("can see that the file is gone", func() {
						result, err = client.ListFiles()
						Ω(err).ShouldNot(HaveOccurred())
						Ω(result).ShouldNot(ContainElement("someother-file.csv"))
					})

					It("returns false when deleting a non-existant files", func() {
						status, err = client.DeleteFile("someother-file.csv")
						Ω(err).ShouldNot(HaveOccurred())
						Ω(status).Should(BeFalse())
					})
				})

				Context("with invalid connection details", func() {
					BeforeEach(func() {
						unsetEnv()
					})

					Context("when listing files", func() {
						BeforeEach(func() {
							result, err = client.ListFiles()
						})
						It("throws an error", func() {
							Ω(err).Should(HaveOccurred())
						})
					})

					Context("when deleting a file", func() {
						BeforeEach(func() {
							status, err = client.DeleteFile("doesntmatter")
						})
						It("throws an error", func() {
							Ω(err).Should(HaveOccurred())
						})
					})

					Context("when uploading a file", func() {
						BeforeEach(func() {
							remoteFilePath, err = client.UploadFile("doesntmatter", "doesntmatter")
						})
						It("throws an error", func() {
							Ω(err).Should(HaveOccurred())
						})
					})

					Context("when downloading a file", func() {
						BeforeEach(func() {
							tempDir, err = ioutil.TempDir("", "iaas-uploading-files")
							localFilePath, err = client.GetFile("doesntmatter", tempDir)
						})
						It("throws an error", func() {
							Ω(err).Should(HaveOccurred())
						})
					})
				})
			})
		})
	})
})
