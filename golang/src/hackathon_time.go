package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"time"
)

type HackathonTimeResponse struct {
	ID            string `json:"$id"`
	RequestTime   string `json:"requestTime"`
	HackathonTime string `json:"hackathonTime"`
}

func getCurrentHackathonTime(token string) (string, error) {
	time.Sleep(80 * time.Millisecond)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	body, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "GET", "/times/HackathonTimeForNow", headers, nil, nil)
	if err != nil {
		log.Println("Error on dispatching request. ", err.Error())
		return "", err
	}

	var hackathonTimeResponse HackathonTimeResponse
	if err := json.Unmarshal(body, &hackathonTimeResponse); err != nil {
		log.Println(err)
	}

	return hackathonTimeResponse.HackathonTime, nil
}

func getDateString(hackathonTime string) string {
	return hackathonTime[0:10]
}

func roundToPrevious15Mins(timeString string) string {
	t, _ := time.Parse(time.RFC3339, timeString)
	// Making 4 quarters in an hour
	quarter := (t.Minute() / 15) * 15

	// Round to the nearest quarter
	//if t.Minute()%15 >= 8 {
	//	quarter += 15
	//}

	// If the minute is 60, change to 00 and add 1 hour
	if quarter == 60 {
		t = t.Add(time.Hour)
		quarter = 0
	}

	// Setting the minute to the nearest quarter
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), quarter, 0, 0, t.Location())

	// Returning the time in RFC3339 format
	return t.Format(time.RFC3339)
}

func roundToNext20Seconds(timeString string) string {
	//log.Println("@@@@@@@@@@@")
	//log.Println(timeString)
	t, _ := time.Parse(time.RFC3339, timeString)
	// Making 4 quarters in an hour
	seconds := (t.Second() / 20) * 20

	// Round to the nearest quarter
	//if t.Minute()%15 >= 8 {
	//	quarter += 15
	//}

	// If the minute is 60, change to 00 and add 1 hour
	//if quarter == 60 {
	//	t = t.Add(time.Hour)
	//	quarter = 0
	//}

	// Setting the minute to the nearest quarter
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), seconds, 0, t.Location())
	//log.Println(t.Format(time.RFC3339))
	t = t.Add(time.Second * 40)
	//log.Println(t.Format(time.RFC3339))
	//log.Println("@@@@@@@@@@@")

	// Returning the time in RFC3339 format
	return t.Format(time.RFC3339)
}

func add15Mins(timeString string) string {
	//log.Println("add15Mins", timeString)
	parsed, _ := time.Parse(time.RFC3339, timeString)
	parsed = parsed.Add(time.Minute * 15)
	//log.Println(parsed)
	//log.Println(parsed.Format(time.RFC3339))
	return parsed.Format(time.RFC3339)
}

func add1Minute(timeString string) string {
	//log.Println("add20Seconds", timeString)
	parsed, _ := time.Parse(time.RFC3339, timeString)
	parsed = parsed.Add(time.Minute)
	//log.Println(parsed)
	//log.Println(parsed.Format(time.RFC3339))
	return parsed.Format(time.RFC3339)
}

func getNextDay(dateString string) string {
	parsed, _ := time.Parse("2006-01-02", dateString)
	parsed = parsed.AddDate(0, 0, 1)
	return parsed.Format("2006-01-02")
}

func getHackathonTime(token, realTime string) (string, error) {
	time.Sleep(80 * time.Millisecond)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	params := url.Values{}
	params.Add("realTime", realTime)

	body, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "GET", "/times/HackathonTimeForDateTime", headers, params, nil)
	if err != nil {
		return "", err
	}

	var hackathonTimeResponse HackathonTimeResponse
	if err := json.Unmarshal(body, &hackathonTimeResponse); err != nil {
		return "", err
	}

	return hackathonTimeResponse.HackathonTime, nil
}

func timeDiffSeconds(start string, stop string) float64 {
	parsedStart, _ := time.Parse(time.RFC3339, start)
	parsedStop, _ := time.Parse(time.RFC3339, stop)
	diff := parsedStop.Sub(parsedStart)
	//log.Println(parsedStart, parsedStop, diff)
	return diff.Seconds()
}
