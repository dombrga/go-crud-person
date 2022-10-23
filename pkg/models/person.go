package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type PersonRequest struct {
	FirstName     string `json:"firstName,omitempty" bson:"firstName,omitempty" validate:"required"`
	LastName      string `json:"lastName,omitempty" bson:"lastName,omitempty" validate:"required"`
	Birthdate     string `json:"birthdate,omitempty" bson:"birthdate,omitempty" validate:"required"`
	IsSoftDeleted bool   `json:"isSoftDeleted" bson:"isSoftDeleted"`
}

type Person struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName     string             `json:"firstName,omitempty" bson:"firstName,omitempty"`
	LastName      string             `json:"lastName,omitempty" bson:"lastName,omitempty"`
	Birthdate     string             `json:"birthdate,omitempty" bson:"birthdate,omitempty"`
	IsSoftDeleted bool               `json:"isSoftDeleted" bson:"isSoftDeleted"`
}
