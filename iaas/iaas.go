package iaas

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type IaaSClient interface {
	DeleteFile(remotePath string) (wasPreExisting bool, err error)
	GetFile(remotePath string, localDir string) (downloadedFilePath string, err error)
	ListFiles() (names []string, err error)
	UploadFile(filepath string, target string) (name string, err error)
	AddFileUploadNotification() (wasNewConfiguration bool, err error)
	FileUploadNotification() (isSet bool, err error)
	RemoveFileUploadNotification() (wasPreExisting bool, err error)
}

type AwsClient struct {
	Region       string
	session      *session.Session
	IntegratorId string
	ClientId     string
}

func (client AwsClient) RemoveFileUploadNotification() (wasPreExisting bool, err error) {
	fullConfig, err := client.getUploadNotificationConfiguration()
	if err != nil {
		return
	}
	configs := fullConfig.TopicConfigurations

	foundIndex := -1
	for i, configuration := range configs {
		if *configuration.Id == client.getNotificationId() {
			foundIndex = i
			break
		}
	}

	if foundIndex < 0 {
		log.Println("Upload notifications not found when attempting removal for", client.ClientId)
		return
	}

	wasPreExisting = true

	// dont care about order, swap final element with unwanted element and resize slice 1 smaller
	configs[len(configs)-1], configs[foundIndex] = configs[foundIndex], configs[len(configs)-1]
	configs = configs[:len(configs)-1]

	fullConfig = fullConfig.SetTopicConfigurations(configs)
	_, err = client.putUploadNotificationConfiguration(fullConfig)
	log.Println("Upload notification removed for", client.ClientId)
	return
}

func (client AwsClient) FileUploadNotification() (isSet bool, err error) {

	fullConfig, err := client.getUploadNotificationConfiguration()
	if err != nil {
		return
	}
	configs := fullConfig.TopicConfigurations

	for _, configuration := range configs {
		if *configuration.Id == client.getNotificationId() {
			log.Println("Upload notifications are set for", client.ClientId)
			isSet = true
			break
		}
	}
	log.Println("Upload notifications are not set for", client.ClientId)
	return
}

func (client AwsClient) AddFileUploadNotification() (wasNewConfiguration bool, err error) {

	fullConfig, err := client.getUploadNotificationConfiguration()
	if err != nil {
		return
	}
	configs := fullConfig.TopicConfigurations

	for _, configuration := range configs {
		if *configuration.Id == client.getNotificationId() {
			log.Println("Upload notifications were already configured when adding for", client.ClientId)
			return
		}
	}
	wasNewConfiguration = true
	accountId, err := client.getAccountId()

	thisConfig := &s3.TopicConfiguration{
		Events: []*string{
			aws.String("s3:ObjectCreated:*"),
		},
		TopicArn: aws.String("arn:aws:sns:" + client.Region + ":" + accountId + ":S3NotifierTopic"),
		Filter: &s3.NotificationConfigurationFilter{
			Key: &s3.KeyFilter{
				FilterRules: []*s3.FilterRule{
					{
						Name:  aws.String("Prefix"),
						Value: aws.String(client.getNotificationPrefix()),
					},
				},
			},
		},
		Id: aws.String(client.getNotificationId()),
	}

	configs = append(configs, thisConfig)
	fullConfig = fullConfig.SetTopicConfigurations(configs)
	_, err = client.putUploadNotificationConfiguration(fullConfig)
	log.Println("Upload notifications added for", client.ClientId)
	return
}

func (client AwsClient) ListFiles() (names []string, err error) {
	names = []string{}

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := s3.New(session)

	params := &s3.ListObjectsInput{
		Bucket: aws.String(client.bucketName()),
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

func (client AwsClient) DeleteFile(remotePath string) (wasPreExisting bool, err error) {

	files, err := client.ListFiles()
	if err != nil {
		return
	}
	if !arrayContains(files, remotePath) {
		return
	}

	wasPreExisting = true

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := s3.New(session)

	targetFile := client.ClientId + "/" + remotePath

	params := &s3.DeleteObjectInput{
		Bucket: aws.String(client.bucketName()),
		Key:    aws.String(targetFile),
	}

	_, err = svc.DeleteObject(params)

	if err != nil {
		log.Println(err.Error())
		return
	}

	return
}

func (client AwsClient) UploadFile(filepath string, targetName string) (name string, err error) {

	targetFile := client.ClientId + "/" + targetName

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

	encType := "AES256"

	params := &s3.PutObjectInput{
		Bucket:               aws.String(client.bucketName()),
		Key:                  aws.String(targetFile),
		Body:                 fileReader,
		ServerSideEncryption: &encType,
	}

	_, err = svc.PutObject(params)

	if err != nil {
		log.Println(err.Error())
		return
	}
	name = targetName
	log.Println("File", filepath, "uploaded to", targetName)
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
			Bucket: aws.String(client.bucketName()),
			Key:    aws.String(targetFile),
		})
	if err != nil {
		log.Println("Failed to download file", err)
		return
	}

	log.Println("Downloaded ", remotePath, "to", file.Name(), "size", numBytes, "bytes")

	return
}

func (client AwsClient) getUploadNotificationConfiguration() (config *s3.NotificationConfiguration, err error) {

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := s3.New(session)

	params := &s3.GetBucketNotificationConfigurationRequest{
		Bucket: aws.String(client.bucketName()),
	}
	resp, err := svc.GetBucketNotificationConfiguration(params)

	if err != nil {
		log.Println(err.Error())
		return
	}
	config = resp
	return
}

func (client AwsClient) putUploadNotificationConfiguration(config *s3.NotificationConfiguration) (result bool, err error) {

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := s3.New(session)

	params := &s3.PutBucketNotificationConfigurationInput{
		Bucket: aws.String(client.bucketName()),
		NotificationConfiguration: config,
	}
	_, err = svc.PutBucketNotificationConfiguration(params)

	if err != nil {
		log.Println(err.Error())
		return
	}
	result = true
	return
}

func (client AwsClient) getAccountId() (accountId string, err error) {

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := iam.New(session)

	params := &iam.GetUserInput{}
	resp, err := svc.GetUser(params)

	if err != nil {
		log.Println(err.Error())
		return
	}

	userArn := *resp.User.Arn
	// ARNs look like arn:aws:iam::ACCOUNTID:user/USERID
	// Note the double colon after iam, which makes account ID element 4 rather than 3
	accountId = strings.Split(userArn, ":")[4]
	return
}

func (client AwsClient) getNotificationId() string {
	return "S3ObjectCreatedSNS-" + client.IntegratorId + "-" + client.ClientId
}

func (client AwsClient) getNotificationPrefix() string {
	return client.ClientId + "/INPUT"
}

func (client AwsClient) bucketName() string {
	return "/" + client.IntegratorId
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

func arrayContains(haystack []string, needle string) bool {
	for _, hay := range haystack {
		if needle == hay {
			return true
		}
	}
	return false
}
