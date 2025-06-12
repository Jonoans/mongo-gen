package main

import (
	"github.com/jonoans/mongo-gen/examples/output"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Example() {
	output.Initialise(
		output.Config{
			DatabaseName: "MY_DATABASE",
		},
		options.Client().ApplyURI("mongodb://mongodb0.example.com:27017"),
	)

	var models []output.Model
	if err := output.FindMany(&models, bson.M{"field": "value"}); err != nil {
		// handle error
	}
}
