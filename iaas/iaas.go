package iaas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"os"
	"path"
	"strings"
)

type IaaSClient interface {
	ListFiles() (names []string, err error)
	UploadFile(filepath string) (name string, err error)
	GetFile(remotePath string, localDir string) (downloadedFilePath string, err error)
}

type AwsClient struct {
	Region       string
	session      *session.Session
	IntegratorId string
	ClientId     string
}

func (client AwsClient) ListFiles() (names []string, err error) {
	names = []string{}

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := s3.New(session)

	params := &s3.ListObjectsInput{
		Bucket: aws.String("/" + client.IntegratorId),
		Prefix: aws.String(client.ClientId),
	}
	resp, err := svc.ListObjects(params)

	if err != nil {
		log.Println(err.Error())
		return
	}

	for _, key := range resp.Contents {
		objectPath := *key.Key
		objectName := strings.Join(strings.Split(objectPath, "/")[1:], "/")
		names = append(names, objectName)
	}
	return
}

func (client AwsClient) UploadFile(filepath string) (name string, err error) {

	name = path.Base(filepath)
	targetFile := client.ClientId + "/" + name

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := s3.New(session)

	fileReader, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer fileReader.Close()

	params := &s3.PutObjectInput{
		Bucket: aws.String("/" + client.IntegratorId),
		Key:    aws.String(targetFile),
		Body:   fileReader,
	}

	_, err = svc.PutObject(params)

	if err != nil {
		log.Println(err.Error())
		return
	}

	return
}

func (client AwsClient) GetFile(remotePath string, localDir string) (downloadedFilePath string, err error) {

	name := path.Base(remotePath)
	downloadedFilePath = path.Join(localDir, name)
	targetFile := client.ClientId + "/" + name

	if !exists(localDir) {
		err = os.MkdirAll(localDir, 0722)
		if err != nil {
			return
		}
	}

	file, err := os.Create(downloadedFilePath)
	if err != nil {
		log.Fatal("Failed to create file", err)
	}
	defer file.Close()

	session, err := client.connect()
	if err != nil {
		return
	}

	downloader := s3manager.NewDownloader(session)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String("/" + client.IntegratorId),
			Key:    aws.String(targetFile),
		})
	if err != nil {
		log.Println("Failed to download file", err)
		return
	}

	log.Println("Downloaded file", file.Name(), numBytes, "bytes")

	return
}

func (client AwsClient) connect() (sess *session.Session, err error) {
	sess, err = session.NewSession(&aws.Config{
		Region: aws.String(client.Region),
	})

	if err != nil {
		log.Println("Failed to connect:", err)
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		log.Println("Credentials not set:", err)
	}

	return
}

func exists(path string) bool {
	_, err := os.Stat(path)

	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}
