package controller_test

type IaaSClientMock struct {
	Credentials map[string]string
	FilesList   []string
	FileName    string
	FilePath    string
	Success     bool
	Err         error
}

func (client IaaSClientMock) ListFiles() (names []string, err error) {
	if client.Err != nil {
		return nil, client.Err
	}
	return client.FilesList, nil
}

func (client IaaSClientMock) UploadFile(filepath string, targetName string) (name string, err error) {
	if client.Err != nil {
		return "", client.Err
	}
	return client.FileName, nil
}

func (client IaaSClientMock) GetFile(remotePath string, localDir string) (downloadedFilePath string, err error) {
	if client.Err != nil {
		return "", client.Err
	}
	return client.FilePath, nil
}

func (client IaaSClientMock) DeleteFile(remotePath string) (result bool, err error) {
	if client.Err != nil {
		return false, client.Err
	}
	return client.Success, nil
}

func (client IaaSClientMock) AddFileUploadNotification() (wasNewConfiguration bool, err error) {
	if client.Err != nil {
		return false, client.Err
	}
	return client.Success, nil
}

func (client IaaSClientMock) FileUploadNotification() (isSet bool, err error) {
	if client.Err != nil {
		return false, client.Err
	}
	return client.Success, nil
}

func (client IaaSClientMock) RemoveFileUploadNotification() (wasPreExisting bool, err error) {
	if client.Err != nil {
		return false, client.Err
	}
	return client.Success, nil
}

func (client IaaSClientMock) CreateClientUser() (credentials map[string]string, err error) {
	if client.Err != nil {
		return nil, client.Err
	}
	return client.Credentials, nil
}

func (client IaaSClientMock) DeleteClientUser() (err error) {
	return
}
