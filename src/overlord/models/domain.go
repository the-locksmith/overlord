package models

import (
	"gopkg.in/mgo.v2/bson"
	"net"
	"time"
)

type Domain struct {
	ID          bson.ObjectId `storm:"id,increment,index" json:"id"`
	CreatedAt   time.Time     `json:"created_at"`
	Host        string        `storm:"unique,index" json:"host,omitempty"`
	IPAddresses []net.IP      `json:"ip_addresses,omitempty"`
	Headers     string        `json:"headers"`
}
