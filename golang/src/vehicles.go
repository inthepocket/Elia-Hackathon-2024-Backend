package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Vehicle struct {
	ID       string  `bson:"_id"`
	Model    string  `bson:"model"`
	Version  string  `bson:"version"`
	Capacity int     `bson:"capacity"`
	RangeKm  int     `bson:"range"`
	Ean      string  `bson:"ean"`
	KmPerKwh float64 `bson:"kmPerKwh"`
	Reward   float64 `bson:"reward"`
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

func getVehicleData(mongo *mongo.Client, ean string, accessToken string) (VehicleResponse, error) {
	vehicle := getVehicleByEan(mongo, ean)

	assetState, err := getCurrentAssetState(accessToken, ean)

	if err != nil {
		assetState = nil
		return VehicleResponse{}, err
	}

	log.Println("Asset state:", assetState)

	assetSessionsLast24h, _ := getAssetSessionsForDay(accessToken, ean, time.Now().Format(time.RFC3339))
	// if err != nil {
	// 	assetSessionsLast24h = []Session{}
	// 	return
	// }

	vehicleResponse := VehicleResponse{
		Metadata:            vehicle,
		CurrentState:        *assetState,
		SessionsLast24hours: assetSessionsLast24h,
	}

	return vehicleResponse, nil
}

func addReward(mongo *mongo.Client, ean string, reward float64) {
	vehicle := getVehicleByEan(mongo, ean)
	//assetState, err := getCurrentAssetState(accessToken, ean)

	coll := mongo.Database("api").Collection("vehicles")
	ctx := context.TODO()
	coll.UpdateOne(ctx, bson.M{"ean": ean}, bson.D{
		{"$set", bson.D{
			{"reward", vehicle.Reward + reward},
		}}})

}
func getAndStoreVehicleSessions(mongo *mongo.Client, accessToken string, ean string) {
	assetSessionsLast24h, err := getAssetSessionsForDay(accessToken, ean, time.Now().Format(time.RFC3339))
	if err != nil {
		assetSessionsLast24h = []Session{}
	}

	coll := mongo.Database("api").Collection("sessions")
	ctx := context.TODO()

	for _, session := range assetSessionsLast24h {
		filter := bson.M{"startState": session.StartState}
		update := bson.M{"$set": session}

		result, err := coll.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
		if err != nil {
			log.Println("Error upserting session: ", err)
		}
		log.Println("Upserted session: ", result)
	}

}

func getAllVehiclesAndStoreSessions(mongo *mongo.Client, accessToken string) {
	for {
		vehicles := getAllVehicles(mongo)

		for _, vehicle := range vehicles {
			getAndStoreVehicleSessions(mongo, accessToken, vehicle.Ean)
		}

		time.Sleep(time.Minute)
	}
}
