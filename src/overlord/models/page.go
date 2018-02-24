package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Page struct {
	ID        bson.ObjectId `storm:"id,increment,index" json:"id"`
	Path      string        `json:"path,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	URL       string        `json:"url"`
	Scheme    string        `json:"scheme,omitempty"`
	Body      string        `json:"body,omitempty"`
	RawQuery  string        `json:"raw_query"`
	Headers   string        `json:"headers"`
}
