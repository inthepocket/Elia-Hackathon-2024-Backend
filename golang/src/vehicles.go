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
	Metadata          Vehicle
	CurrentState      AssetState
	SessionsLast5Days []Session
}

func getAllVehicles(mongo *mongo.Client) ([]Vehicle, error) {
	vehicles := []Vehicle{}

	coll := mongo.Database("api").Collection("vehicles")
	ctx := context.TODO()

	// Find all vehicles
	cursor, err := coll.Find(ctx, bson.M{})

	if err != nil {
		return []Vehicle{}, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var vehicle Vehicle
		if err = cursor.Decode(&vehicle); err != nil {
			panic(err)
		}

		vehicles = append(vehicles, vehicle)

	}

	return vehicles, nil
}

func getVehicleByEan(mongo *mongo.Client, ean string) (Vehicle, error) {
	vehicle := Vehicle{}

	coll := mongo.Database("api").Collection("vehicles")
	ctx := context.TODO()

	// Find vehicle by EAN
	err := coll.FindOne(ctx, bson.M{"ean": ean}).Decode(&vehicle)

	if err != nil {
		return Vehicle{}, err
	}

	return vehicle, nil
}

func getVehicleData(mongo *mongo.Client, ean string, accessToken string) (VehicleResponse, error) {
	vehicle, err := getVehicleByEan(mongo, ean)
	if err != nil {
		return VehicleResponse{}, err
	}

	log.Println("Vehicle:", vehicle)

	assetState, err := getCurrentAssetState(accessToken, ean)
	if err != nil {
		assetState = nil
		return VehicleResponse{}, err
	}

	log.Println("Asset state:", assetState)

	now := time.Now()
	sessions := []Session{}

	for i := 0; i < 5; i++ {
		date := now.Add(time.Duration(-i*72) * time.Minute)

		// log.Println("Getting sessions for", date.Format(time.RFC3339))
		assetSessions, err := getAssetSessionsForDay(accessToken, ean, date.Format(time.RFC3339))

		if err != nil {
			log.Println("Error getting asset sessions: ", err)
			continue
		}

		sessions = append(sessions, assetSessions...)
	}

	vehicleResponse := VehicleResponse{
		Metadata:          vehicle,
		CurrentState:      *assetState,
		SessionsLast5Days: sessions,
	}

	return vehicleResponse, nil
}

func addReward(mongo *mongo.Client, ean string, reward float64) error {
	vehicle, err := getVehicleByEan(mongo, ean)
	if err != nil {
		return err
	}

	coll := mongo.Database("api").Collection("vehicles")
	ctx := context.TODO()
	coll.UpdateOne(ctx, bson.M{"ean": ean}, bson.D{
		{"$set", bson.D{
			{"reward", vehicle.Reward + reward},
		}}})

	return nil
}

func setLastHourMax(mongo *mongo.Client, ean string, lastHourMax string) error {
	coll := mongo.Database("api").Collection("vehicles")
	ctx := context.TODO()
	coll.UpdateOne(ctx, bson.M{"ean": ean}, bson.D{
		{"$set", bson.D{
			{"lastHourMax", lastHourMax},
		}}})

	return nil
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
		time.Sleep(time.Minute)

		vehicles, err := getAllVehicles(mongo)

		if err != nil {
			continue
		}

		for _, vehicle := range vehicles {
			getAndStoreVehicleSessions(mongo, accessToken, vehicle.Ean)
		}
	}
}
