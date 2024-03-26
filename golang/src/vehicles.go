package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Vehicle struct {
	_id      string `bson:"_id"`
	model    string `bson:"model"`
	version  string `bson:"version"`
	capacity int    `bson:"capacity"`
	rangeKm  int    `bson:"range"`
	ean      string `bson:"ean"`
}

func getAllVehicles(mongo *mongo.Client) []Vehicle {
	vehicles := []Vehicle{}

	coll := mongo.Database("api").Collection("vehicles")
	ctx := context.TODO()

	// Find all vehicles
	cursor, err := coll.Find(ctx, bson.M{})

	if err != nil {
		panic(err)
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var vehicle Vehicle
		if err = cursor.Decode(&vehicle); err != nil {
			panic(err)
		}

		vehicles = append(vehicles, vehicle)

	}

	return vehicles
}
