package iaas_test

import (
	. "github.com/dhrapson/sched-load/iaas"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"os"
	"path"
)

var (
	client          IaaSClient
	accessKeyId     string
	secretAccessKey string
	region          string
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

var _ = Describe("The IaaS Client", func() {

	BeforeSuite(func() {
		accessKeyId = os.Getenv("TEST_AWS_ACCESS_KEY_ID")
		secretAccessKey = os.Getenv("TEST_AWS_SECRET_ACCESS_KEY")
		Ω(accessKeyId).ShouldNot(BeEmpty(), "You must set TEST_AWS_ACCESS_KEY_ID environment variable")
		Ω(secretAccessKey).ShouldNot(BeEmpty(), "You must set TEST_AWS_SECRET_ACCESS_KEY environment variable")
	})

	JustBeforeEach(func() {
		client = AwsClient{IntegratorId: "test-integrator", ClientId: "test-client", Region: region}
	})

	Describe("Interacting with AWS", func() {

		var (
			result []string
			err    error
		)

		Describe("listing files", func() {
			JustBeforeEach(func() {
				result, err = client.ListFiles()
			})

			Context("connecting with valid connection details", func() {
				BeforeEach(func() {
					setEnv()
				})

				AfterEach(func() {
					unsetEnv()
				})

				It("connects correctly & can see the contents", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(result).Should(ContainElement("test-folder/"))
				})
			})

			Context("with invalid connection details", func() {

				It("throws an error", func() {
					Ω(err).Should(HaveOccurred())
				})
			})
		})

		Describe("uploading files", func() {
			var (
				uploadedFile string
				tempDir      string
			)

			JustBeforeEach(func() {
				uploadedFile, err = client.UploadFile("fixtures/test-file.csv", "someother-file.csv")
			})

			Context("connecting with valid connection details", func() {
				BeforeEach(func() {
					setEnv()
					tempDir, err = ioutil.TempDir("", "iaas-uploading-files")
					Ω(err).ShouldNot(HaveOccurred())
				})

				AfterEach(func() {
					unsetEnv()
				})

				It("connects correctly & uploads the file", func() {
					Ω(err).ShouldNot(HaveOccurred())
					Ω(uploadedFile).Should(Equal("someother-file.csv"))
					_, err := client.GetFile(uploadedFile, tempDir)
					Ω(err).ShouldNot(HaveOccurred())
					contents, err := ioutil.ReadFile(path.Join(tempDir, "someother-file.csv"))
					Ω(err).ShouldNot(HaveOccurred())
					expectedContents, err := ioutil.ReadFile("fixtures/test-file.csv")
					Ω(err).ShouldNot(HaveOccurred())
					Ω(contents).Should(Equal(expectedContents))
				})
			})

			Context("with invalid connection details", func() {

				It("throws an error", func() {
					Ω(err).Should(HaveOccurred())
				})
			})
		})
	})
})
