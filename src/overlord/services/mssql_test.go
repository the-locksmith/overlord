package services_test

import (
	"models"

	"testing"
)

func TestScanMssql(t *testing.T) {
	s := models.Service{Ip: "127.0.0.1", Port: 1433, Protocol: "mssql", Username: "sa", Password: ""}
	t.Log(plugins.ScanMssql(s))
}
