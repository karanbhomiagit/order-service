package models

import "gopkg.in/mgo.v2/bson"

type Order struct {
	ID       bson.ObjectId `bson:"_id" json:"id"`
	Distance int           `bson:"distance" json:"distance"`
	Status   string        `bson:"status" json:"status"`
}

type OrderRequest struct {
	Origin      []string `json:"origin"`
	Destination []string `json:"destination"`
}
