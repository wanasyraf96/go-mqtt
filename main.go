package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

var (
	client     mqtt.Client
	clientOnce sync.Once
	clientLock sync.Mutex
)

func createMqttClient() {
	broker_host := os.Getenv("MQTT_URL")
	if broker_host == "" {
		broker_host = "localhost"
	}

	broker_port, _ := strconv.Atoi(os.Getenv("MQTT_PORT"))
	if broker_port == 0 {
		broker_port = 1883
	}
	mqtt_protocol := os.Getenv("MQTT_PROTOCOL")
	if mqtt_protocol == "" {
		mqtt_protocol = "tcp"
	}

	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf(`%s://%s:%d/mqtt`, mqtt_protocol, broker_host, broker_port))
	mqtt_protocol_version, _ := strconv.Atoi(os.Getenv("MQTT_PROTOCOL_VERSION"))
	if mqtt_protocol_version == 0 {
		opts.SetProtocolVersion(uint(mqtt_protocol_version))
		mqtt_protocol_version = 3
	}

	// Create an MQTT client options

	// Set the callback function for connection lost
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("Connection lost: %v", err)
		tryReconnect()
	}

	// Create an MQTT client
	client = mqtt.NewClient(opts)

	// Connect to the MQTT broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("Failed to connect to MQTT broker: %v", token.Error())
		tryReconnect()
	}
}

func tryReconnect() {
	clientLock.Lock()
	defer clientLock.Unlock()

	if client.IsConnected() {
		return
	}

	log.Println("Attempting to reconnect to MQTT broker...")
	for !client.IsConnected() {
		createMqttClient()
		time.Sleep(5 * time.Second)
	}
	log.Println("Reconnected to MQTT broker")
}

func getClient() mqtt.Client {
	clientOnce.Do(createMqttClient)
	return client
}

type mqttReq struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

func main() {
	// Load .env
	godotenv.Load(".env")

	// Define an HTTP handler function for the /mqtt route
	http.HandleFunc("/mqtt", func(w http.ResponseWriter, r *http.Request) {
		// parse request into mqttReq type
		var req mqttReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println(err)
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		// Get the MQTT client instance
		client := getClient()

		// Publish an MQTT message to the "test" topic
		token := client.Publish(req.Topic, 0, false, req.Payload)
		token.Wait()

		fmt.Fprintln(w, "MQTT message published")
	})

	// Start the HTTP server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Server is running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
