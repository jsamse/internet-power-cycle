package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

type dotServer struct {
	IP         string
	ServerName string
}

var dotServers = []dotServer{
	{"1.1.1.1:853", "cloudflare-dns.com"},
	{"8.8.8.8:853", "dns.google"},
	{"9.9.9.9:853", "dns.quad9.net"},
	{"185.228.168.9:853", "security-filter-dns.cleanbrowsing.org"},
	{"94.140.14.14:853", "dns.adguard.com"},
}

func hasInternetConnection() bool {
	timeout := 3 * time.Second
	for _, server := range dotServers {
		conn, err := tls.DialWithDialer(
			&net.Dialer{
				Timeout: timeout,
			},
			"tcp",
			server.IP,
			&tls.Config{
				ServerName:         server.ServerName,
				InsecureSkipVerify: false,
			},
		)

		if err == nil {
			conn.Close()
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
			fmt.Println("Waiting for 5 minutes before checking again...")
			time.Sleep(2 * time.Minute)
		}
		time.Sleep(3 * time.Minute)
	}
}
