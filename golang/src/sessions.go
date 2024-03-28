package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChargePeriod struct {
	StartTime  string
	EndTime    string
	SocAtStart float32
	SocAtEnd   float32
	ChargedKwh float32
}

type Session struct {
	rewardsForSession float32
	StartState        *AssetState `json:"StartState"`
	EndState          *AssetState `json:"EndState"`
	ChargePeriods     []ChargePeriod
}

func getAndStoreCurrentSessions(token string, mongo *mongo.Client) {
	for {
		time.Sleep(time.Second * 10)

		vehicles, _ := getAllVehicles(mongo)
		date := time.Now()
		for _, vehicle := range vehicles {
			time.Sleep(time.Second * 5)

			log.Println("Getting current session for vehicle", vehicle.Ean)
			assetSessions, err := getAssetSessionsForDay(token, vehicle.Ean, date.Format(time.RFC3339))

			if err != nil {
				log.Println("Error getting asset sessions: ", err)
				continue
			}

			coll := mongo.Database("api").Collection("sessions")

			for _, session := range assetSessions {
				filter := bson.M{"startState.ean": session.StartState.Ean, "startState.stateTime": session.StartState.StateTime}
				update := bson.M{"$set": session}
				opts := options.Update().SetUpsert(true)
				ctx := context.TODO()
				_, err := coll.UpdateOne(ctx, filter, update, opts)
				if err != nil {
					log.Println("Error upserting session: ", err.Error())
				}
			}

		}
	}
}

func getMostRecentVehicleSession(mongo *mongo.Client, ean string) (*Session, error) {
	coll := mongo.Database("api").Collection("sessions")
	ctx := context.TODO()
	opts := options.FindOne().SetSort(bson.D{{"startState.stateTime", -1}})
	filter := bson.M{"startState.ean": ean}
	var session Session
	err := coll.FindOne(ctx, filter, opts).Decode(&session)
	if err != nil {
		log.Println("Error getting most recent session: ", err.Error())
		return nil, err
	}

	return &session, nil
}

func getSessionsForVehicle(mongo *mongo.Client, ean string) ([]Session, error) {
	coll := mongo.Database("api").Collection("sessions")
	ctx := context.TODO()
	filter := bson.M{
		"startstate.ean": ean,
	}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		log.Println("Error getting sessions for vehicle: ", err.Error())
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []Session
	for cursor.Next(ctx) {
		var session Session
		err := cursor.Decode(&session)
		if err != nil {
			log.Println("Error decoding session: ", err.Error())
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}
