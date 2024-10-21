package bootlrSearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"encore.dev/rlog"
)

// =================
// SEARCH SERVICES
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
			Model:       "gpt-4o-mini",
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

	rlog.Info("BOOTLR SEARCH", "OpenAI responseData ==> ", string(responseData))

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
	productsPerPage := 40

	url := fmt.Sprintf("https://real-time-product-search.p.rapidapi.com/search?q=%s&country=se&language=sv&limit=%d", searchQuery, productsPerPage)
	
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("X-RapidAPI-Key", RAPIDAPI_KEY)
	req.Header.Add("X-RapidAPI-Host", "real-time-product-search.p.rapidapi.com")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	rlog.Info("BOOTLR SEARCH",
		"RAPID BODY ==> ", string(body))

	var response map[string]interface{}
	json.Unmarshal(body, &response)

	data := response["data"].(map[string]interface{})
	searchResults := data["products"].([]interface{})

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