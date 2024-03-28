package main

import (
	"log"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var eans []string = []string{"541983310278725782", "541792416037150809", "541480930594258894", "541261957970628703", "541560923639711466"}

func steerAssets(token string, mongo *mongo.Client) {

	previousHackathonTime := ""
	currentHackathonTime := ""
	roofPrices := make(map[string]RoofPrices)
	lastKnownSoc := make(map[string]float32)
	carMinChargeLevels := make(map[string]float32)
	carRewards := make(map[string]float32)

	previousMongoDbFlushTime := time.Now()

	for {
		time.Sleep(time.Second * 2)
		log.Println("################## ######## ##################")
		log.Println("################## Starting ##################")
		log.Println("################## ######## ##################")

		// Get date of current day
		previousHackathonTime = currentHackathonTime
		currentHackathonTime, err := getCurrentHackathonTime(token)
		if err != nil || currentHackathonTime == "" {
			log.Println("###### ABORTING No hackathon time available", err)
			continue
		}
		currentDateString := getDateString(currentHackathonTime)

		// Get day ahead time of new day
		dayAheadPricesJson := getDayAheadPrices(token, currentDateString, currentHackathonTime)

		// Find roof price of cheapest surface
		for _, ean := range eans {
			state, err := getCurrentAssetState(token, ean, currentHackathonTime)
			if state.Connected == false {
				continue

			}

			var expectedKwhToCharge int = 100.0
			if err != nil {
				log.Println("Error in getting asset state, expecting default 100 kWh")
			} else {
				expectedKwhToCharge = int(60 - state.Soc)
			}

			roofPrice, err := calculateRoofPricePerQuarter(dayAheadPricesJson, expectedKwhToCharge, expectedKwhToCharge, 1.1)
			if err != nil {
				log.Println(err)
			} else {
				roofPrices[ean] = roofPrice
				//setLastHourMax(getMongoClient(), ean, fmt.Sprintf("%d:00:00", int(roofPrice.LastHourMax)))
			}

		}
		log.Println("### roofPrices ", roofPrices)

		// Get real-time price
		currentRealTimePrice, err := getRealTimePrice(token, currentHackathonTime)
		if err != nil || math.Abs(currentRealTimePrice) < 0.001 {
			log.Println("###### ABORTING No real time price available ", err)
			continue
		}
		log.Println("### real time price:", currentRealTimePrice)

		// Calculate rewards per car
		cars := getActiveCars(token)
		log.Println("###", "Cars", cars)
		for _, car := range cars {
			//log.Println(car.consumptionKwSincePreviousTime, timeDiffSeconds(previousHackathonTime, currentHackathonTime), currentRealTimePrice)
			if currentRealTimePrice > 0 {
				continue
			}
			reward := car.consumptionKwSincePreviousTime * float32(timeDiffSeconds(previousHackathonTime, currentHackathonTime)*currentRealTimePrice/1000/3600*-1)
			//log.Println(car.Ean, "Reward: ", reward)

			var currentCarReward float32 = 0
			if val, ok := carRewards[car.Ean]; ok {
				currentCarReward = val
			}
			carRewards[car.Ean] = currentCarReward + reward*0.25
		}
		log.Println("### carRewards", carRewards)

		// Calculate lastKnownSoc per car
		for _, car := range cars {
			lastKnownSocCar := lastKnownSoc[car.Ean]
			currentSoc := car.Soc

			if currentSoc < lastKnownSocCar {
				carMinChargeLevels[car.Ean] = (lastKnownSocCar - currentSoc) * 1.5
			}
		}
		log.Println("### lastKnownSoc", lastKnownSoc)

		// Calculate charge per car
		charge := make(map[string]bool)
		for _, ean := range eans {
			charge[ean] = roofPrices[ean].RoofMax > float32(currentRealTimePrice)
		}
		log.Println("### charges", charge)

		// Send all steeringRequests
		// Add 2 minutes to compensate for long calculation time
		parsedCurrentHackathonTime, _ := time.Parse(time.RFC3339, currentHackathonTime)
		parsedCurrentHackathonTime = parsedCurrentHackathonTime.Add(time.Second * 120)
		//updatedCurrentHackathonTime := parsedCurrentHackathonTime.Format(time.RFC3339)

		//steeringRequest(token, updatedCurrentHackathonTime, cars, charge, carMinChargeLevels)

		if time.Now().Sub(previousMongoDbFlushTime).Seconds() > 120 {
			log.Println("###### Flushing to MongoDb")
			for _, car := range cars {
				if carReward, ok := carRewards[car.Ean]; ok {
					addReward(mongo, car.Ean, float64(carReward))
				}
				carRewards[car.Ean] = 0
			}
			previousMongoDbFlushTime = time.Now()

		}

	}

}

func steerBattery(token string) {
	for {
		time.Sleep(time.Second * 5)

		hackathonTime, err := getCurrentHackathonTime(token)
		if err != nil {
			log.Println("Error on getting hackathon time. ", err.Error())
			continue
		}
		realTimePrice, err := getRealTimePrice(token, hackathonTime)
		if err != nil || math.Abs(realTimePrice) < 0.001 {
			log.Println("###### No real time price available ", err)
			continue
		}

		log.Println("Steering Battery...")
		log.Println("Hackathon Time: ", hackathonTime)
		log.Println("Real Time Price: ", realTimePrice)

		if realTimePrice < 0 {
			log.Println("Real time price is negative, charge battery")
			steeringRequestBattery(token, hackathonTime, true)
		} else {
			log.Println("Real time price is positive, do not charge battery")
			steeringRequestBattery(token, hackathonTime, false)
		}

	}

}
