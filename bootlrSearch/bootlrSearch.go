package bootlrSearch

import (
	"encoding/json"
	"net/http"
	"encore.dev/rlog"
)

var secrets struct {
	OPENAI_KEY string
	OPENAI_ORG string
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
	searchHistory, err := RetreiveSearchRequestBody(write, req)
	if err != nil {
		http.Error(write, "Error retreiving request body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rlog.Info("BOOTLR SEARCH", "REQ Body (searchHistory) ==> ", searchHistory)

	searchQuery, err := TranslateMessagesToSearchQuery(searchHistory)
	if err != nil {
		http.Error(write, "Error translating message to search term: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shoppingResults, err := GetShoppingResults(searchQuery, "se")
	if err != nil {
		http.Error(write, "Error getting shopping results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: get as many amazon products as possible (up to 20, but 10 works) without throttling 429
	// consider also a retry method if throttled to wait 1 second and try again up to 2 or 3 times
	// TODO: Data translation service -> amazon to rapid struct or vice versa

	write.Header().Set("Content-Type", "application/json")
	response := SearchResponse{
		SearchQuery:     searchQuery,
		ShoppingResults: shoppingResults,
	}
	json.NewEncoder(write).Encode(response)
}

