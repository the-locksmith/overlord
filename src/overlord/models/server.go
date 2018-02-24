package models

import (
	"gopkg.in/mgo.v2/bson"
	"net"
	"time"
)

type Server struct {
	ID           bson.ObjectId `storm:"id,increment,index" json:"id"`
	CreatedAt    time.Time     `json:"created_at"`
	IPAddress    net.IP        `storm:"index" json:"ip_address"`
	Host         string        `storm:"unique,index" json:"host,omitempty"`
	OpenTCPPorts []int         `json:"open_tcp_ports,omitempty"`
	OpenUDPPorts []int         `json:"open_udp_ports,omitempty"`
	Services     []Service     `json:"services,omitempty"`
	// Use other tools to get these
	//WebApplication string
	//Banner         string
}
