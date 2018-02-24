package services_test

import (
	"models"

	"testing"
)

func TestScanMongodb(t *testing.T) {
	s := models.Service{Ip: "127.0.0.1", Port: 27017, Username: "test", Password: "test", Protocol: "mongodb"}
	t.Log(plugins.ScanMongodb(s))
}
