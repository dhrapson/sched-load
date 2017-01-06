package controller

import (
	"errors"
	"github.com/dhrapson/sched-load/iaas"
	"io/ioutil"
	"os"
	"path"
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

func (controller Controller) UploadFile(filePath string) (result string, err error) {

	result = "error"
	targetFile := path.Base(filePath)
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

func (controller Controller) SetSchedule(interval string) (result bool, err error) {
	result = false
	targetFile := interval + "_SCHEDULE"
	var tempFile *os.File
	if tempFile, err = ioutil.TempFile("", "set-schedule"); err != nil {
		return
	}

	var fileName string
	fileName, err = controller.Client.UploadFile(tempFile.Name(), targetFile)
	result = (fileName == targetFile)
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
