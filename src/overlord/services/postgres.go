package services

import (
	_ "github.com/lib/pq"

	"models"

	"database/sql"
	"fmt"
)

func ScanPostgres(service models.Service) (err error, result models.ScanResult) {
	result.Service = service

	dataSourceName := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=%v", service.Username,
		service.Password, service.Ip, service.Port, "postgres", "disable")
	db, err := sql.Open("postgres", dataSourceName)

	if err == nil {
		defer db.Close()
		err = db.Ping()
		if err == nil {
			result.Result = true
		}
	}
	return err, result
}
