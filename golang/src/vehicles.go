package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Vehicle struct {
	ID       string  `bson:"_id"`
	Model    string  `bson:"model"`
	Version  string  `bson:"version"`
	Capacity int     `bson:"capacity"`
	RangeKm  int     `bson:"range"`
	Ean      string  `bson:"ean"`
	KmPerKwh float64 `bson:"kmPerKwh"`
}

type VehicleResponse struct {
	Metadata            Vehicle
	CurrentState        AssetState
	SessionsLast24hours []Session
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

func getVehicleByEan(mongo *mongo.Client, ean string) Vehicle {
	vehicle := Vehicle{}

	coll := mongo.Database("api").Collection("vehicles")
	ctx := context.TODO()

	// Find vehicle by EAN
	err := coll.FindOne(ctx, bson.M{"ean": ean}).Decode(&vehicle)

	if err != nil {
		panic(err)
	}

	return vehicle
}
