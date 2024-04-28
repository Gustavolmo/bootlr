package bootlrSearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	/* "strings" */
	g "github.com/serpapi/google-search-results-golang"
)

var secrets struct {
	OPENAI_KEY string
	OPENAI_ORG string
	SERPAPI_KEY string
	IPLOCATIONAPI_KEY string
}

type MessageHistoryItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
    Model       string   							`json:"model"`
    Messages    []MessageHistoryItem  `json:"messages"`
    Temperature float64  							`json:"temperature"`
}

type OpenAIResponse struct {
    Choices []struct {
        Message struct {
            Content string `json:"content"`
        } `json:"message"`
    } `json:"choices"`
}

type SearchResponse struct {
	SearchQuery     string        `json:"searchQuery"`
	ShoppingResults []interface{} `json:"shoppingResults"`
}

//encore:api public method=POST raw path=/bootlr-search
func BootlrSearch(write http.ResponseWriter, req *http.Request) {
	messageHistory, err := RetreiveSearchRequestBody(write, req)
	if err != nil {
		http.Error(write, "Error retreiving request body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	searchQuery, err := TranslateMessagesToSearchQuery(messageHistory)
	if err != nil {
		http.Error(write, "Error translating message to search term: "+err.Error(), http.StatusInternalServerError)
		return
	}

	/* reqLocation, err := getRequestLocation(req)
	if err != nil {
		http.Error(write, "Error getting location results: "+err.Error(), http.StatusInternalServerError)
		return
	} */

	shoppingResults, err := GetShoppingResults(searchQuery/* , reqLocation */)
	if err != nil {
		http.Error(write, "Error getting shopping results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	write.Header().Set("Content-Type", "application/json")
	response := SearchResponse{
		SearchQuery:     searchQuery,
		ShoppingResults: shoppingResults,
	}
	json.NewEncoder(write).Encode(response)
}

// =================
// HELPER FUNCTIONS
// =================

func RetreiveSearchRequestBody(write http.ResponseWriter, req *http.Request) ([]MessageHistoryItem, error) {
	if req.Method != http.MethodPost {
		http.Error(write, "Method not allowed", http.StatusMethodNotAllowed)
	}

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(write, "Error reading request body: "+err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	var messageHisotry []MessageHistoryItem
	if err := json.Unmarshal(reqBody, &messageHisotry); err != nil {
		http.Error(write, "Error unmarshaling JSON: "+err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	return messageHisotry, nil
}

func TranslateMessagesToSearchQuery(messageHistory []MessageHistoryItem) (string, error) {
	OPENAI_KEY := secrets.OPENAI_KEY
	OPENAI_ORG := secrets.OPENAI_ORG
	OPENAI_URL := "https://api.openai.com/v1/chat/completions"

	openAiRequestBody := OpenAIRequest{
			Model:       "gpt-3.5-turbo",
			Messages:    messageHistory,
			Temperature: 1.2,
	}

	openAiRequestBodyBytes, err := json.Marshal(openAiRequestBody)
	if err != nil {
			return "", err
	}

	client := http.Client{}
	openAiRequest, err := http.NewRequest("POST", OPENAI_URL, bytes.NewBuffer(openAiRequestBodyBytes))
	if err != nil {
			return "", err
	}
	openAiRequest.Header.Set("Content-Type", "application/json")
	openAiRequest.Header.Set("OpenAI-Organization", OPENAI_ORG)
	openAiRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", OPENAI_KEY))

	response, err := client.Do(openAiRequest)
	if err != nil {
			return "", err
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
			return "", err
	}

	var openAIResponse OpenAIResponse
	if err := json.Unmarshal(responseData, &openAIResponse); err != nil {
		return "", err
	}

	searchQuery := openAIResponse.Choices[0].Message.Content
	return searchQuery, nil
}

func GetShoppingResults(query string/* , reqLocation string */) ([]interface{}, error) {
	SERPAPI_KEY := secrets.SERPAPI_KEY

	parameter := map[string]string{
    "engine": "google_shopping",
    "q": query,
    "api_key": SERPAPI_KEY,
		"gl": "SE",
  }

	search := g.NewGoogleSearch(parameter, SERPAPI_KEY)
  results, err := search.GetJSON()
	if err != nil {
		return nil, err
	}
  shoppingResults := results["shopping_results"].([]interface{})
	return shoppingResults, nil
}

/* func getRequestLocation(req *http.Request) (string, error){
	IPLOCATIONAPI_KEY := secrets.IPLOCATIONAPI_KEY

	ip := strings.Split(req.RemoteAddr, ":")[0]
	ipGeoLocationUrl := fmt.Sprintf("https://api.ipgeolocation.io/ipgeo?apiKey=%s&ip=%s", IPLOCATIONAPI_KEY, ip)

	geoReq, err := http.NewRequest("GET", ipGeoLocationUrl, nil)
	if err != nil {
		return "", err
	}

	geoResponse, err := http.DefaultClient.Do(geoReq)
	if err != nil {
		return "", err
	}
	defer geoResponse.Body.Close()
	
	var geoData map[string]interface{}
	err = json.NewDecoder(geoResponse.Body).Decode(&geoData)
	if err != nil {
		return "", err
	}

	countryCode, ok := geoData["country_code2"].(string)
	if !ok {
		return "", nil
	}

	return countryCode, nil
} */