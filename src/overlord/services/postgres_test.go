package services_test

import (
	"models"

	"testing"
)

func TestScanPostgres(t *testing.T) {
	s := models.Service{Ip: "127.0.0.1", Port: 5432, Protocol: "postgres", Username: "postgres", Password: ""}
	t.Log(plugins.ScanPostgres(s))
}
