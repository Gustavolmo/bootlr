package bootlrSearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var secrets struct {
	OPENAI_KEY string
	OPENAI_ORG string
	SERPAPI_KEY string
	RAPIDAPI_KEY string
}

type GeoData struct {
	Country struct {
		ISOCode string `json:"iso_code"`
	} `json:"country"`
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

	shoppingResults, err := GetShoppingResults(searchQuery, "se")
	if err != nil {
		http.Error(write, "Error getting shopping results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: get as many amazon products as possible (up to 60, but 10 works) without throttling 429
	// consider also a retry method if throttled to wait 1 second and try again up to 2 or 3 times

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

func GetShoppingResults(query string, reqLocation string) ([]interface{}, error) {
	RAPIDAPI_KEY := secrets.RAPIDAPI_KEY
	searchQuery := query

	url := fmt.Sprintf("https://real-time-product-search.p.rapidapi.com/search?q=%s&country=se&language=en", searchQuery)
	
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("X-RapidAPI-Key", RAPIDAPI_KEY)
	req.Header.Add("X-RapidAPI-Host", "real-time-product-search.p.rapidapi.com")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var response map[string]interface{}
	json.Unmarshal(body, &response)
	searchResults := response["data"].([]interface{})

	return searchResults, nil
}

//================
// PARKED SERVICES
//================

// TODO: ENABLE WHEN SEARCH PER LOCATION IS AVAILABLE
/* func getRequestCountryCode(req *http.Request) (string, error) {
	ip := req.Header.Get("Cf-Connecting-Ip")
	ip = strings.Split(ip, ",")[0]
	
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
		ip = strings.Split(ip, ",")[0]
	}

	if ip == "" {
		ip = req.RemoteAddr
		ip = strings.Split(ip, ",")[0]
	}

	url := fmt.Sprintf("https://api.findip.net/%s/?token=6bad0e5471f8429a9028ef23448e6e02", ip)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting geo info:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error decoding geo info:", err)
		return "", err
	}

	var geoData GeoData
	json.Unmarshal(body, &geoData)

	return geoData.Country.ISOCode, nil
} */