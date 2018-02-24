package services_test

import (
	"testing"

	"models"
)

func TestScanMysql(t *testing.T) {
	service := models.Service{Ip: "127.0.0.1", Port: 3306, Protocol: "mysql", Username: "root", Password: "123456"}
	t.Log(plugins.ScanMysql(service))
}
