package services

import (
	"github.com/jlaffaye/ftp"

	"models"

	"fmt"
)

func ScanFtp(s models.Service) (err error, result models.ScanResult) {
	result.Service = s
	conn, err := ftp.DialTimeout(fmt.Sprintf("%v:%v", s.Ip, s.Port), vars.TimeOut)
	if err == nil {
		err = conn.Login(s.Username, s.Password)
		if err == nil {
			defer conn.Logout()
			result.Result = true
		}
	}
	return err, result
}
