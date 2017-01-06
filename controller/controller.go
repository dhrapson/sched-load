package controller

import (
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
