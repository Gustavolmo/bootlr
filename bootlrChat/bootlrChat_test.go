package bootlrChat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Run tests using `encore test`, which compiles the Encore app and then runs `go test`.
// It supports all the same flags that the `go test` command does.
// You automatically get tracing for tests in the local dev dash: http://localhost:9400
// Learn more: https://encore.dev/docs/develop/testing

func TestBootlrSearch(t *testing.T) {
	// Mocking the request body
	requestBody := `{"messages":[{"role":"system","content":"sys_message1"}, {"role":"user","content":"user_message2"}]}`

	// Mocking the request
	req, err := http.NewRequest("POST", "/bootlr-chat", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	// Mocking the response writer
	resRecorder := httptest.NewRecorder()

	// Mocking the mocker
	mocker := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mocking the response
		response := ChatResponse{
			ResponseText:     "chat_response_text",
			ProductReference: []string{"product1", "product2", "product3", "product4"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Making the request to the mock handler
	mocker.ServeHTTP(resRecorder, req)

	// Check the status code
	if status := resRecorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Decode the response body
	var response ChatResponse
	err = json.NewDecoder(resRecorder.Body).Decode(&response)
	if err != nil {
		t.Errorf("error decoding response body: %v", err)
	}

	// Check the content of the response
	expectedResponseText := "chat_response_text"
	if response.ResponseText != expectedResponseText {
		t.Errorf("unexpected search query: got %v want %v", response.ResponseText, expectedResponseText)
	}
	
	ProductReference := response.ProductReference
	ExpectedFourthProduct := "product4"
	if len(ProductReference) != 4 || ProductReference[3] != ExpectedFourthProduct {
		t.Errorf("unexpected shopping result: got %v want %v", ProductReference[3], ExpectedFourthProduct)
	}
}
