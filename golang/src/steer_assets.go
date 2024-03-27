package main

import (
	"log"
	"math"
	"time"
)

func steerAssets(token string) {

	previousHackathonTime := ""
	currentHackathonTime := ""
	currentDateString := ""
	var roofPrices RoofPrices
	lastKnownSoc := make(map[string]float32)
	carMinChargeLevels := make(map[string]float32)
	carRewards := make(map[string]float32)

	previousMongoDbFlushTime := time.Now()

	for {
		time.Sleep(time.Second * 2)
		log.Println("### Starting")

		// Get date of current day
		previousHackathonTime = currentHackathonTime
		newHackathonTime, err := getCurrentHackathonTime(token)
		if err != nil || newHackathonTime == "" {
			log.Println("###### No hackathon time available", err)
			continue
		}
		currentHackathonTime = newHackathonTime
		log.Println("### Current time:", currentHackathonTime)
		newDateString := getDateString(currentHackathonTime)

		// If start of new day (or need not set)
		if newDateString != currentDateString {
			log.Println("### Starting new day", newDateString)

			currentDateString = newDateString

			// Calculate or assume new need
			//need := 300 // kWh

			// Get day ahead time of new day
			dayAheadPricesJson := getDayAheadPrices(token, currentDateString)

			// Find roof of cheapest surface
			var expectedKwhToCharge int = 100.0
			roofPrices = calculateRoofPricePerQuarter(dayAheadPricesJson, expectedKwhToCharge, expectedKwhToCharge, 1.1)
		}

		log.Println("### roofPrices:", roofPrices.RoofComfort, roofPrices.RoofMax)
		// Get real-time price
		currentRealTimePrice, err := getRealTimePrice(token, currentHackathonTime)
		if err != nil || math.Abs(currentRealTimePrice) < 0.001 {
			log.Println("###### No real time price available ", err)
			continue
		}
		log.Println("### real time price:", currentRealTimePrice)

		// If real-time price < roof => charge
		//if roofPrices.RoofComfort > float32(currentRealTimePrice) {
		//	log.Println("CHARGE WITH NO REWARD")
		//}

		cars := getActiveCars(token)
		log.Println("###", "Cars", cars)
		for _, car := range cars {
			//log.Println(car.consumptionKwSincePreviousTime, timeDiffSeconds(previousHackathonTime, currentHackathonTime), currentRealTimePrice)
			reward := car.consumptionKwSincePreviousTime * float32(timeDiffSeconds(previousHackathonTime, currentHackathonTime)*currentRealTimePrice/1000/3600)
			log.Println(car.Ean, "Reward: ", reward)

			var currentCarReward float32 = 0
			if val, ok := carRewards[car.Ean]; ok {
				currentCarReward = val
			}
			if reward < 0 {
				panic("Negative reward")
			}
			carRewards[car.Ean] = currentCarReward + reward*0.25
		}

		for _, car := range cars {
			lastKnownSocCar := lastKnownSoc[car.Ean]
			currentSoc := car.Soc

			if currentSoc < lastKnownSocCar {
				carMinChargeLevels[car.Ean] = (lastKnownSocCar - currentSoc) * 1.5
			}
		}

		if roofPrices.RoofMax > float32(currentRealTimePrice) {
			log.Println("CHARGE")
		} else {
			log.Println("DO NOT CHARGE")
		}
		steeringRequest(token, currentHackathonTime, cars, roofPrices.RoofMax > float32(currentRealTimePrice), carMinChargeLevels)
		log.Println(carRewards)

		if time.Now().Sub(previousMongoDbFlushTime).Seconds() > 120 {
			log.Println("###### Flushing to MongoDb")
			for _, car := range cars {
				if carReward, ok := carRewards[car.Ean]; ok {
					addReward(getMongoClient(), car.Ean, float64(carReward))
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
