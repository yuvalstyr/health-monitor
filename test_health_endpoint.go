package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"health-monitor/internal/handlers"
)

func main() {
	// Test the health endpoint with different scenarios
	fmt.Println("Testing health endpoint...")

	// Test 1: Health endpoint with default version
	fmt.Println("\n1. Testing health endpoint with default configuration:")
	testHealthEndpoint("http://localhost:3000/health")

	// Test 2: Health endpoint with custom version
	fmt.Println("\n2. Testing health endpoint with custom version:")
	os.Setenv("APP_VERSION", "2.1.0-test")
	testHealthEndpoint("http://localhost:3000/health")

	fmt.Println("\nHealth endpoint testing completed!")
}

func testHealthEndpoint(url string) {
	client := &http.Client{Timeout: 5 * time.Second}
	
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ Failed to make request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status Code: %d\n", resp.Status)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	var healthResp handlers.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		fmt.Printf("❌ Failed to decode response: %v\n", err)
		return
	}

	// Pretty print the response
	respJSON, _ := json.MarshalIndent(healthResp, "", "  ")
	fmt.Printf("Response:\n%s\n", string(respJSON))

	// Validate response structure
	if healthResp.Status == "" {
		fmt.Println("❌ Missing status field")
		return
	}
	if healthResp.Version == "" {
		fmt.Println("❌ Missing version field")
		return
	}
	if healthResp.Timestamp.IsZero() {
		fmt.Println("❌ Missing or invalid timestamp")
		return
	}
	if healthResp.Migrations == nil {
		fmt.Println("❌ Missing migrations field")
		return
	}

	fmt.Println("✅ Health endpoint response is valid")
}