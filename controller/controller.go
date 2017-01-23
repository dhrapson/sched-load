package controller

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/dhrapson/sched-load/iaas"
)

type Controller struct {
	Client iaas.IaaSClient
}

func (controller Controller) Status() (string, error) {
	_, err := controller.Client.ListFiles()
	if err != nil {
		return "error", err
	}
	return "connected", nil
}

func (controller Controller) CreateClientUser() (credentials iaas.IaaSCredentials, err error) {
	credentials, err = controller.Client.CreateClientUser()
	return
}

func (controller Controller) DeleteClientUser(force bool) (wasPreExisting bool, err error) {
	wasPreExisting, err = controller.Client.DeleteClientUser(force)
	return
}
func (controller Controller) ListDataFiles() (result []string, err error) {

	var fileNames []string
	if fileNames, err = controller.Client.ListFiles(); err != nil {
		return
	}

	for _, fileName := range fileNames {
		// if its a key within INPUT/, and not INPUT/ itself
		if strings.Index(fileName, "INPUT/") == 0 && len(fileName) > 6 {
			result = append(result, fileName)
		}
	}

	return
}

func (controller Controller) ImmediateDataFileCollectionStatus() (status bool, err error) {
	status, err = controller.Client.FileUploadNotification()
	return
}

func (controller Controller) EnableImmediateDataFileCollection() (wasNewlyEnabled bool, err error) {
	wasNewlyEnabled, err = controller.Client.AddFileUploadNotification()
	return
}

func (controller Controller) DisableImmediateDataFileCollection() (removedExisting bool, err error) {
	removedExisting, err = controller.Client.RemoveFileUploadNotification()
	return
}

func (controller Controller) UploadDataFile(filePath string) (result string, err error) {

	result = "error"
	targetFile := "INPUT/" + path.Base(filePath)
	var fileName string
	if fileName, err = controller.Client.UploadFile(filePath, targetFile); err != nil {
		return
	}

	var fileNames []string
	if fileNames, err = controller.Client.ListFiles(); err != nil {
		return
	}

	if arrayContains(fileNames, fileName) {
		result = fileName
	} else {
		err = errors.New("Unable to find uploaded file" + fileName)
	}
	return
}

func (controller Controller) DeleteDataFile(filePath string) (result bool, err error) {
	targetFile := "INPUT/" + filePath
	return controller.Client.DeleteFile(targetFile)
}

func (controller Controller) RemoveSchedule() (previouslySet bool, err error) {
	return controller.Client.DeleteFile("DAILY_SCHEDULE")
}

func (controller Controller) SetSchedule(interval string) (result bool, err error) {
	result = false
	targetFile := interval + "_SCHEDULE"
	var tempFile *os.File
	tempFile, err = ioutil.TempFile("", "set-schedule")
	defer tempFile.Close()
	if err != nil {
		return
	}

	var fileName string
	fileName, err = controller.Client.UploadFile(tempFile.Name(), targetFile)
	result = (fileName == targetFile)
	return
}

func (controller Controller) GetSchedule() (result string, err error) {
	result = "ERROR"
	fileName := "DAILY_SCHEDULE"
	var fileNames []string
	if fileNames, err = controller.Client.ListFiles(); err != nil {
		return
	}

	if arrayContains(fileNames, fileName) {
		result = "DAILY"
	} else {
		result = "NONE"
	}
	return
}
func arrayContains(haystack []string, needle string) bool {
	for _, hay := range haystack {
		if needle == hay {
			return true
		}
	}
	return false
}
