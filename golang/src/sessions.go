package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
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
	Trivia            string
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
					continue
				}
				log.Println("/// boomerise_it")
				headers := map[string]string{}
				headers["Content-Type"] = "application/json"
				params := url.Values{}

				if session.EndState == nil {
					log.Println("/// boomerise_it -- session.EndState == nil")
					session.Trivia = "Charging up 60 kWh - that's like fueling up a '60s VW Beetle for a spin around Woodstock!"
					// continue
				}

				// total charged kwh
				totalChargedKwh := int32(0)
				for _, chargePeriod := range session.ChargePeriods {
					totalChargedKwh += int32(chargePeriod.ChargedKwh)
				}

				totalChargedKwh = int32(60)

				dataJson := fmt.Sprintf("{\"ean\": \"%s\", \"state_time\": \"%s\", \"energy_kwh\": %d}",
					session.StartState.Ean,
					session.StartState.StateTime,
					totalChargedKwh)

				log.Println("/// boomerise_it -- dataJson START")
				log.Println(dataJson)
				log.Println("/// boomerise_it -- dataJson END")

				body, err := makeRequest(os.Getenv("STEERING_PYTHON_URI"), "POST", "/boomerise_it", headers, params, bytes.NewBuffer([]byte(dataJson)))
				if err != nil {
					log.Fatal("Error on dispatching request. ", err.Error())
				}
				log.Println(string(body))
				session.Trivia = string(body)
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
