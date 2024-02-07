package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
)

// Event represents a web page visit event
type Event struct {
	URL       string `json:"url"`
	VisitorID string `json:"visitorId"`
	Timestamp int64  `json:"timestamp"`
}

// Session represents a session of web page visits by a visitor
type Session struct {
	Duration  int64    `json:"duration"`
	Pages     []string `json:"pages"`
	StartTime int64    `json:"startTime"`
}

func main() {
	// Fetch the input data
	inputDataURL := "https://candidate.hubteam.com/candidateTest/v3/problem/dataset?userKey=ce939a6e782e5d75bd1149391016"
	response, err := http.Get(inputDataURL)
	if err != nil {
		fmt.Println("Error fetching input data:", err)
		return
	}
	defer response.Body.Close()

	var data struct {
		Events []Event `json:"events"`
	}
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		fmt.Println("Error decoding input data:", err)
		return
	}

	// Group events by visitor ID
	eventsByVisitor := make(map[string][]Event)
	for _, event := range data.Events {
		eventsByVisitor[event.VisitorID] = append(eventsByVisitor[event.VisitorID], event)
	}

	// Step 3: Sort events within each group by timestamp
	for _, events := range eventsByVisitor {
		sort.Slice(events, func(i, j int) bool {
			return events[i].Timestamp < events[j].Timestamp
		})
	}

	// Identify sessions and create sessions
	sessionsByUser := make(map[string][]Session)
	for visitorID, events := range eventsByVisitor {
		sessions := make([]Session, 0)
		currentSession := Session{StartTime: events[0].Timestamp}
		tempTimestamp := events[0].Timestamp
		for i := 0; i < len(events); i++ {
			event := events[i]
			if event.Timestamp-tempTimestamp <= 600000 || len(currentSession.Pages) == 0 {
				currentSession.Pages = append(currentSession.Pages, event.URL)
				currentSession.Duration = event.Timestamp - currentSession.StartTime
				tempTimestamp = event.Timestamp
			} else {
				sessions = append(sessions, currentSession)
				currentSession = Session{StartTime: event.Timestamp, Pages: []string{event.URL}}
				tempTimestamp = event.Timestamp
			}
		}
		sessions = append(sessions, currentSession)
		sessionsByUser[visitorID] = sessions
	}

	// Format the output data
	outputData := struct {
		SessionsByUser map[string][]Session `json:"sessionsByUser"`
	}{
		SessionsByUser: sessionsByUser,
	}

	// Send the formatted output data via HTTP POST request
	postURL := "https://candidate.hubteam.com/candidateTest/v3/problem/result?userKey=ce939a6e782e5d75bd1149391016"
	jsonData, err := json.Marshal(outputData)
	if err != nil {
		fmt.Println("Error encoding output data:", err)
		return
	}

	response, err = http.Post(postURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending data:", err)
		return
	}
	defer response.Body.Close()

	// Check the response status
	if response.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		fmt.Println("Success: Data sent successfully.", bodyString)
	} else {
		fmt.Println("Error:", response.Status)
	}
}
