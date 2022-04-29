package codegen

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionFunc func(ctx mongo.SessionContext) error

type BaseModel struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
}

func (m *BaseModel) GetID() interface{} {
	return m.ID
}

func (m *BaseModel) SetID(id interface{}) {
	m.ID = id.(primitive.ObjectID)
}
