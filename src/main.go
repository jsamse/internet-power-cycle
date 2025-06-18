package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var dnsServers = []string{
	"https://www.cloudflarestatus.com/",
	"https://www.githubstatus.com/",
	"https://google.com",
}

func hasInternetConnection() bool {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	for _, server := range dnsServers {
		resp, err := client.Get(server)
		if err == nil {
			resp.Body.Close()
			return true
		}
	}
	return false
}

func turnOffShellyPlug(ip string) error {
	url := fmt.Sprintf("http://%s/rpc/Switch.Set", ip)
	body := map[string]any{
		"id": 0,
		"on": false,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	ip := os.Getenv("SHELLY_IP")
	if ip == "" {
		fmt.Println("SHELLY_IP environment variable is not set.")
		return
	}

	for {
		retries := 3
		for retries > 0 {
			if hasInternetConnection() {
				break
			}
			retries--
			fmt.Println("No internet connection detected. Retrying...")
			time.Sleep(10 * time.Second)
		}
		if retries == 0 {
			fmt.Println("No internet connection detected. Turning off Shelly plug...")
			err := turnOffShellyPlug(ip)
			if err != nil {
				fmt.Println("Error turning off Shelly plug:", err)
				continue
			}
			fmt.Println("Waiting for 2 minutes to let devices boot...")
			time.Sleep(2 * time.Minute)
		}
		fmt.Println("Waiting for 3 minutes before checking internet connection again...")
		time.Sleep(3 * time.Minute)
	}
}
