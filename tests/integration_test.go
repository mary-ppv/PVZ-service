package tests

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// sendRequest is a helper function to send an HTTP request and return the response
func sendRequest(t *testing.T, method string, url string, token string, payload interface{}) (*http.Response, []byte) {
	var body bytes.Buffer
	// Encode the payload into JSON if it exists
	if payload != nil {
		err := json.NewEncoder(&body).Encode(payload)
		assert.NoError(t, err) // Ensure no error during encoding
	}

	// Create a new HTTP request
	req, err := http.NewRequest(method, url, &body)
	assert.NoError(t, err) // Ensure no error during request creation

	// Set headers for the request
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token) // Add authorization token if provided
	}

	// Send the request using an HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err) // Ensure no error during request execution

	// Read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)  // Ensure no error during reading the response
	defer resp.Body.Close() // Close the response body after reading

	return resp, respBody // Return the response and body
}

// getDummyToken is a helper function to obtain a token via the /dummyLogin endpoint
func getDummyToken(t *testing.T, serverURL string, role string) string {
	payload := map[string]string{"role": role} // Prepare the payload with the desired role
	resp, respBody := sendRequest(t, "POST", serverURL+"/dummyLogin", "", payload)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // Ensure the login was successful

	var result map[string]string
	err := json.Unmarshal(respBody, &result) // Parse the response body to extract the token
	assert.NoError(t, err)                   // Ensure no error during unmarshalling

	return result["token"] // Return the token from the response
}

// TestFullFlow simulates a full user flow: creating a PVZ, starting a reception, adding products, and closing the reception
func TestFullFlow(t *testing.T) {
	// Create a test server to mock API endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dummyLogin":
			// Simulate dummy login
			var req map[string]string
			json.NewDecoder(r.Body).Decode(&req)
			if req["role"] == "moderator" || req["role"] == "employee" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		case "/pvz":
			// Simulate PVZ creation
			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":               "test-pvz-id",
					"city":             "Москва",
					"registrationDate": time.Now().Format(time.RFC3339),
				})
			}
		case "/receptions":
			// Simulate reception creation
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":       "test-reception-id",
				"dateTime": time.Now().Format(time.RFC3339),
				"pvzId":    "test-pvz-id",
				"status":   "in_progress",
			})
		case "/products":
			// Simulate product addition
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          "test-product-id",
				"dateTime":    time.Now().Format(time.RFC3339),
				"type":        "электроника",
				"receptionId": "test-reception-id",
			})
		case "/pvz/test-pvz-id/close_last_reception":
			// Simulate closing the last reception
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":       "test-reception-id",
				"dateTime": time.Now().Format(time.RFC3339),
				"pvzId":    "test-pvz-id",
				"status":   "close",
			})
		default:
			// Handle unknown endpoints
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close() // Close the test server after the test

	serverURL := server.URL

	// Step 1: Get tokens for moderator and employee roles
	moderatorToken := getDummyToken(t, serverURL, "moderator") // Get token for moderator
	employeeToken := getDummyToken(t, serverURL, "employee")   // Get token for employee

	// Step 2: Create a new PVZ (Pickup Point)
	pvzPayload := map[string]string{"city": "Москва"} // Prepare payload for PVZ creation
	resp, respBody := sendRequest(t, "POST", serverURL+"/pvz", moderatorToken, pvzPayload)
	assert.Equal(t, http.StatusCreated, resp.StatusCode) // Ensure PVZ creation was successful

	var pvzResult map[string]interface{}
	err := json.Unmarshal(respBody, &pvzResult) // Parse the response to extract PVZ details
	assert.NoError(t, err)                      // Ensure no error during unmarshalling

	id, ok := pvzResult["id"].(string)                 // Extract the PVZ ID from the response
	assert.True(t, ok, "Expected 'id' to be a string") // Ensure the ID is a string
	pvzID := id

	// Step 3: Start a new reception for the created PVZ
	receptionPayload := map[string]string{"pvzId": pvzID} // Prepare payload for reception creation
	resp, _ = sendRequest(t, "POST", serverURL+"/receptions", employeeToken, receptionPayload)
	assert.Equal(t, http.StatusCreated, resp.StatusCode) // Ensure reception creation was successful

	// Step 4: Add 50 products to the current reception
	for i := 0; i < 50; i++ {
		productPayload := map[string]string{
			"type":  "электроника", // Type of product
			"pvzId": pvzID,         // Associate product with the PVZ
		}
		resp, _ = sendRequest(t, "POST", serverURL+"/products", employeeToken, productPayload)
		assert.Equal(t, http.StatusCreated, resp.StatusCode) // Ensure product addition was successful
	}

	// Step 5: Close the last reception
	closeReceptionURL := serverURL + "/pvz/" + pvzID + "/close_last_reception" // Construct URL for closing reception
	resp, _ = sendRequest(t, "POST", closeReceptionURL, employeeToken, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // Ensure reception closure was successful
}
