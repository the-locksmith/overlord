package services_test

import (
	"models"

	"testing"
)

func TestScanRedis(t *testing.T) {
	s := models.Service{Ip: "127.0.0.1", Port: 6379, Password: "test"}
	t.Log(plugins.ScanRedis(s))
}
