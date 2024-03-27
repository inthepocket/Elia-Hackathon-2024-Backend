package main

import (
	"log"
	"time"
)

func steerAssets(token string) {

	currentDateString := ""
	var roofPrices RoofPrices

	for {
		// Get date of current day
		currentHackathonTime := getCurrentHackathonTime(token)
		newDateString := getDateString(currentHackathonTime)
		log.Println("### Current time:", currentHackathonTime)

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
			roofPrices = calculateRoofPricePerQuarter(dayAheadPricesJson, expectedKwhToCharge, expectedKwhToCharge, 2.0)
		}

		log.Println("### roofPrices:", roofPrices.RoofComfort, roofPrices.RoofMax)
		// Get real-time price
		currentRealTimePrice := getRealTimePrice(token, currentHackathonTime)
		log.Println("### real time price:", currentRealTimePrice)

		// If real-time price < roof => charge
		//if roofPrices.RoofComfort > float32(currentRealTimePrice) {
		//	log.Println("CHARGE WITH NO REWARD")
		//}

		cars := getActiveCars(token)
		log.Println("###", "Cars", cars)

		if roofPrices.RoofMax > float32(currentRealTimePrice) {
			//if true {
			log.Println("CHARGE")
		} else {
			log.Println("DO NOT CHARGE")
		}
		steeringRequest(token, currentHackathonTime, cars, roofPrices.RoofMax > float32(currentRealTimePrice))

		time.Sleep(time.Second * 2)

	}

}

func steerBattery(token string) {
	for {
		hackathonTime := getCurrentHackathonTime(token)
		realTimePrice := getRealTimePrice(token, hackathonTime)
		batteryState, err := getCurrentAssetState(token, "541657038024211911")
		if err != nil {
			log.Println("Error getting battery state")
		}

		isBatteryFullyCharged := batteryState.Soc == 100
		isBatteryDischarging := batteryState.Production > 0

		if realTimePrice < 0 && !isBatteryFullyCharged && !isBatteryDischarging {
			log.Println("Real time price is negative, charge battery")
			steeringRequestBattery(token, hackathonTime, true)
		} else {
			log.Println("Real time price is positive, do not charge battery")
			steeringRequestBattery(token, hackathonTime, false)
		}

		time.Sleep(time.Second * 2)

	}

}
