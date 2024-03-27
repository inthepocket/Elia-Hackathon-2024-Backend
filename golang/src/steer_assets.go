package main

import "log"

func steerAssets(token string) {

	currentDateString := ""

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
			calculateRoofPricePerQuarter(dayAheadPricesJson)
		}

		// Get real-time price

		// If real-time price < roof => charge

	}

}
