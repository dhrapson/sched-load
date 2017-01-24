package iaas

import (
	"errors"
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

type IaaSAccountDetails map[string]string

type IaaSCredentials interface {
	String() string
	Map() map[string]string
}

type IaaSClient interface {
	DeleteFile(remotePath string) (wasPreExisting bool, err error)
	GetFile(remotePath string, localDir string) (downloadedFilePath string, err error)
	ListFiles() (names []string, err error)
	UploadFile(filepath string, target string) (name string, err error)
	AddFileUploadNotification() (wasNewConfiguration bool, err error)
	FileUploadNotification() (isSet bool, err error)
	RemoveFileUploadNotification() (wasPreExisting bool, err error)
	CreateClientUser() (credentials IaaSCredentials, err error)
	DeleteClientUser(force bool) (wasPreExisting bool, err error)
	AccountDetails() (details IaaSAccountDetails, err error)
}

type AwsClient struct {
	Region       string
	session      *session.Session
	IntegratorId string
	ClientId     string
	AccountId    string
}

type AwsCredentials struct {
	AccessKeyId     string
	SecretAccessKey string
}

func (creds AwsCredentials) String() (output string) {
	return "AccessKeyId: " + creds.AccessKeyId + ", SecretAccessKey: " + creds.SecretAccessKey
}

func (creds AwsCredentials) Map() map[string]string {
	m := make(map[string]string)
	m["AccessKeyId"] = creds.AccessKeyId
	m["SecretAccessKey"] = creds.SecretAccessKey
	return m
}

func (client AwsClient) RemoveFileUploadNotification() (wasPreExisting bool, err error) {

	if err = client.populate(); err != nil {
		return
	}

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

	if err = client.populate(); err != nil {
		return
	}

	fullConfig, err := client.getUploadNotificationConfiguration()
	if err != nil {
		return
	}
	configs := fullConfig.TopicConfigurations

	for _, configuration := range configs {
		if *configuration.Id == client.getNotificationId() {
			isSet = true
			break
		}
	}
	if isSet {
		log.Println("Upload notifications are set for", client.ClientId)
	} else {
		log.Println("Upload notifications are not set for", client.ClientId)
	}
	return
}

func (client AwsClient) AddFileUploadNotification() (wasNewConfiguration bool, err error) {

	if err = client.populate(); err != nil {
		return
	}

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

	thisConfig := &s3.TopicConfiguration{
		Events: []*string{
			aws.String("s3:ObjectCreated:*"),
		},
		TopicArn: aws.String("arn:aws:sns:" + client.Region + ":" + client.AccountId + ":S3NotifierTopic"),
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

	if err = client.populate(); err != nil {
		return
	}

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

	if err = client.populate(); err != nil {
		return
	}

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

	if err = client.populate(); err != nil {
		return
	}

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

	if err = client.populate(); err != nil {
		return
	}

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

func (client AwsClient) CreateClientUser() (credentials IaaSCredentials, err error) {

	if err = client.populate(); err != nil {
		return
	}

	err = client.createClientUser()
	if err != nil {
		return
	}
	err = client.addClientUserToIntegratorClientGroup()
	if err != nil {
		return
	}
	credentials, err = client.createClientAccessKey()
	log.Println("Created client user account for " + client.ClientId)
	return
}

func (client AwsClient) DeleteClientUser(force bool) (wasPreExisting bool, err error) {

	if err = client.populate(); err != nil {
		return
	}

	wasPreExisting, err = client.clientUserExists()
	if err != nil {
		return
	}

	if force {
		var files []string
		files, err = client.ListFiles()
		if err != nil {
			return
		}
		for _, file := range files {
			_, err = client.DeleteFile(file)
			if err != nil {
				return
			}
		}
	}

	if !wasPreExisting {
		return
	}

	wasInGroup, err := client.isClientUserInIntegratorClientGroup()
	if err != nil {
		return
	}
	if wasInGroup {
		err = client.removeClientUserFromIntegratorClientGroup()
		if err != nil {
			return
		}
	}
	keys, err := client.listClientAccessKeys()
	if err != nil {
		return
	}

	for _, key := range keys {
		err = client.deleteClientAccessKey(key)
		if err != nil {
			return
		}
	}

	err = client.deleteClientUser()
	if err != nil {
		return
	}

	return
}

func (client AwsClient) AccountDetails() (details IaaSAccountDetails, err error) {

	details = map[string]string{}
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
	details["AccountId"] = strings.Split(userArn, ":")[4]

	userPath := strings.Split(userArn, ":")[5]
	pathParts := strings.Split(userPath, "/")

	if len(pathParts) > 3 || pathParts[0] != "user" {
		return nil, errors.New("Unexpected user path: " + userPath)
	}
	if pathParts[1] == "integrator" {
		details["IntegratorId"] = pathParts[2]
		details["ConnectionType"] = "integrator"
	} else {
		details["IntegratorId"] = pathParts[1]
		details["ClientId"] = pathParts[2]
		details["ConnectionType"] = "client"
	}

	return
}

func (client *AwsClient) populate() error {
	if client.IntegratorId == "" || client.AccountId == "" {
		details, err := client.AccountDetails()
		if err != nil {
			return err
		}
		client.AccountId = details["AccountId"]
		client.IntegratorId = details["IntegratorId"]
		mapValue, ok := details["ClientId"]
		if ok {
			if client.ClientId != "" && client.ClientId != mapValue {
				return errors.New("Client ID mismatch: Given client ID " + client.ClientId + " does not match ID for IaaS credentials: " + mapValue)
			}
			client.ClientId = mapValue
		}
	}
	return nil
}

func (client AwsClient) clientUserExists() (exists bool, err error) {
	session, err := client.connect()
	if err != nil {
		return
	}

	svc := iam.New(session)

	params := &iam.GetUserInput{
		UserName: aws.String(client.ClientId),
	}

	_, err = svc.GetUser(params)

	if err != nil {
		if strings.Contains(err.Error(), "AccessDenied") {
			err = nil
		} else {
			log.Println(err)
		}
		return
	}
	exists = true
	return
}

func (client AwsClient) addClientUserToIntegratorClientGroup() (err error) {

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := iam.New(session)

	params := &iam.AddUserToGroupInput{
		GroupName: aws.String(client.integratorClientGroupName()), // Required
		UserName:  aws.String(client.ClientId),                    // Required
	}
	_, err = svc.AddUserToGroup(params)

	if err != nil {
		log.Println(err.Error())
		return
	}

	return
}

func (client AwsClient) isClientUserInIntegratorClientGroup() (isGroupMember bool, err error) {

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := iam.New(session)

	params := &iam.ListGroupsForUserInput{
		UserName: aws.String(client.ClientId), // Required
	}
	resp, err := svc.ListGroupsForUser(params)

	if err != nil {
		log.Println(err.Error())
		return
	}

	for _, iamGroup := range resp.Groups {
		if *iamGroup.GroupName == client.integratorClientGroupName() {
			return true, nil
		}
	}
	return
}

func (client AwsClient) integratorClientGroupName() string {
	return client.IntegratorId + "-client"
}
func (client AwsClient) removeClientUserFromIntegratorClientGroup() (err error) {

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := iam.New(session)

	params := &iam.RemoveUserFromGroupInput{
		GroupName: aws.String(client.integratorClientGroupName()), // Required
		UserName:  aws.String(client.ClientId),                    // Required
	}
	_, err = svc.RemoveUserFromGroup(params)

	if err != nil {
		log.Println(err.Error())
		return
	}

	return
}

func (client AwsClient) createClientAccessKey() (credentials AwsCredentials, err error) {

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := iam.New(session)

	params := &iam.CreateAccessKeyInput{
		UserName: aws.String(client.ClientId),
	}
	resp, err := svc.CreateAccessKey(params)

	if err != nil {
		log.Println(err.Error())
		return
	}

	credentials = AwsCredentials{AccessKeyId: *resp.AccessKey.AccessKeyId,
		SecretAccessKey: *resp.AccessKey.SecretAccessKey}
	return
}

func (client AwsClient) listClientAccessKeys() (names []string, err error) {
	names = []string{}
	session, err := client.connect()
	if err != nil {
		return
	}

	svc := iam.New(session)

	params := &iam.ListAccessKeysInput{
		UserName: aws.String(client.ClientId),
	}
	resp, err := svc.ListAccessKeys(params)

	if err != nil {
		log.Println(err.Error())
		return
	}
	for _, metadata := range resp.AccessKeyMetadata {
		names = append(names, *metadata.AccessKeyId)
	}

	return
}

func (client AwsClient) deleteClientAccessKey(accessKeyId string) (err error) {

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := iam.New(session)

	params := &iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(accessKeyId),
		UserName:    aws.String(client.ClientId),
	}

	_, err = svc.DeleteAccessKey(params)

	if err != nil {
		log.Println(err.Error())
		return
	}
	return
}

func (client AwsClient) deleteClientUser() (err error) {
	session, err := client.connect()
	if err != nil {
		return
	}

	svc := iam.New(session)

	params := &iam.DeleteUserInput{
		UserName: aws.String(client.ClientId),
	}
	_, err = svc.DeleteUser(params)

	if err != nil {
		log.Println(err.Error())
	}
	return
}

func (client AwsClient) createClientUser() (err error) {

	session, err := client.connect()
	if err != nil {
		return
	}

	svc := iam.New(session)

	params := &iam.CreateUserInput{
		UserName: aws.String(client.ClientId), // Required
		Path:     aws.String("/" + client.IntegratorId + "/"),
	}
	_, err = svc.CreateUser(params)

	if err != nil {
		log.Println(err.Error())
		return
	}
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
