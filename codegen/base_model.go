package codegen

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TransactionFunc func(ctx context.Context) error

type BaseModel struct {
	ID bson.ObjectID `bson:"_id,omitempty"`
}

func (m *BaseModel) GetID() interface{} {
	return m.ID
}

func (m *BaseModel) SetID(id interface{}) {
	m.ID = id.(bson.ObjectID)
}
