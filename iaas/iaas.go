package iaas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"strings"
)

type IaaSClient struct {
	Region       string
	session      *session.Session
	IntegratorId string
	ClientId     string
}

func (client IaaSClient) ListFiles() (names []string, err error) {
	names = []string{}

	session, err := client.connect()
	if err != nil {
		log.Println("Failed to connect:", err)
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

func (client IaaSClient) connect() (sess *session.Session, err error) {
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
