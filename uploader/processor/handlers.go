package processor

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type ProcessBody struct {
	File []uint8 `json:"file"`
	Name string  `json:"name"`
}

type Message struct {
	Timestamp *string `json:"timestamp"`
	Type      string  `json:"type"`
}

func (a *Api) StartTaskHandler(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	var taskEvent ProcessBody
	err := decoder.Decode(&taskEvent)

	if err != nil {
		msg := fmt.Sprintf("Error matching the body: %v\n", err)
		log.Print(msg)
		res.WriteHeader(400)
		e := ErrResponse{
			HTTPStatusCode: 400,
			Message:        msg,
		}
		json.NewEncoder(res).Encode(e)
		return
	}

	id, err := a.Processor.Convert(taskEvent.File)
	if err != nil {
		msg := fmt.Sprintf("Error decoding file content: %v\n", err)
		log.Print(msg)
		res.WriteHeader(http.StatusBadRequest)
		e := ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			Message:        msg,
		}
		json.NewEncoder(res).Encode(e)
		return
	}
	res.WriteHeader(201)
	json.NewEncoder(res).Encode(map[string]interface{}{
		"id":  id,
		"url": fmt.Sprintf("ws://%s:%v/socket/%s", a.Address, a.Port, id),
	})
	a.handlers[id] = time.Now()
}

func (a *Api) StartStreamHandler(res http.ResponseWriter, req *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	urlParts := strings.Split(req.URL.Path, "/")
	id := urlParts[len(urlParts)-1]
	log.Printf("Client connected with ID: %s", id)

	data, err := a.Processor.readJSON(id)
	if err != nil {
		log.Printf("File reading error: %v", err)
		return
	}

	currentTimestamp, ok := data[0]["timestamp"].(string)
	if !ok {
		log.Println("Failed to assert timestamp type")
		return
	}

	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	registrationTime := a.handlers[id]
	if time.Now().After(registrationTime.Add(1 * time.Hour)) {
		conn.WriteMessage(8, []byte("1 hour has passed"))
	}

	log.Println("Client connected")

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}
		log.Printf("Received message: %s", message)

		var msg Message
		var parsedTime time.Time
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			conn.WriteMessage(messageType, []byte("faulty message"))
			continue
		}
		if msg.Type != "subscribe" && msg.Type != "next" {
			log.Println("Invalid type")
			break
		}

		record, index, ok := a.Processor.findRecordByTimestamp(data, currentTimestamp)
		if msg.Type == "subscribe" {
			if msg.Timestamp != nil && *msg.Timestamp != "" {
				parsedTime, err = time.Parse(time.RFC3339, *msg.Timestamp)
				if err != nil {
					log.Printf("Error parsing timestamp: %v", err)
				}
			}
			if err == nil {
				newRecord, newIndex, newOk := a.Processor.findRecordByTimestamp(data, parsedTime.String())
				if newOk {
					record = newRecord
					index = newIndex
					ok = newOk
				}
			}
		}
		if !ok {
			conn.WriteMessage(messageType, []byte(fmt.Sprintf("%s", "finished")))
			break
		}

		jsonResponse, _ := json.Marshal(record)

		err = conn.WriteMessage(messageType, []byte(fmt.Sprintf("%s", jsonResponse)))

		if index == len(data)-1 {
			conn.WriteMessage(messageType, []byte(fmt.Sprintf("%s", "finished")))
			break
		}
		currentTimestamp = data[index+1]["timestamp"].(string)

		if err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
	time.Sleep(1 * time.Hour)
	log.Println("Client disconnected")
	return
}
