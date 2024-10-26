package bootlrSearch

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
	requestBody := `{"messages":[{"role":"system","content":"sys_message1"}, {"role":"system","content":"sys_message2"}]}`

	// Mocking the request
	req, err := http.NewRequest("POST", "/bootlr-search", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	// Mocking the response writer
	resRecorder := httptest.NewRecorder()

	// Mocking the mocker
	mocker := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mocking the response
		response := SearchResponse{
			SearchQuery:     "search_query",
			ShoppingResults: []interface{}{
				map[string]interface{}{"product1": "product1_of_many"},
			},
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
	var response SearchResponse
	err = json.NewDecoder(resRecorder.Body).Decode(&response)
	if err != nil {
		t.Errorf("error decoding response body: %v", err)
	}

	// Check the content of the response
	expectedQuery := "search_query"
	if response.SearchQuery != expectedQuery {
		t.Errorf("unexpected search query: got %v want %v", response.SearchQuery, expectedQuery)
	}
	
	shoppingResults := response.ShoppingResults
	expectedProduct := "product1_of_many"
	if len(shoppingResults) != 1 || shoppingResults[0].(map[string]interface{})["product1"] != expectedProduct {
		t.Errorf("unexpected shopping result: got %v want %v", shoppingResults[0].(map[string]interface{})["product1"], expectedProduct)
	}
}
