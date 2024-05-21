package bootlrChat

import (
	"encoding/json"
	"net/http"
	"encore.dev/rlog"
)

var secrets struct {
	OPENAI_KEY string
	OPENAI_ORG string
}

type ChatResponse struct {
	ResponseText     string   `json:"responseText"`
	ProductReference []string `json:"productReference"`
}

type MessageHistoryItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
    Model       		string   									`json:"model"`
    Messages    		[]MessageHistoryItem  		`json:"messages"`
    Temperature 		float64  									`json:"temperature"`
		Response_format OpenaiResponseformat			`json:"response_format"`
}

type OpenaiResponseformat struct {
	Type string `json:"type"`
}

type OpenAIResponse struct {
    Choices []struct {
        Message struct {
            Content string `json:"content"`
        } `json:"message"`
    } `json:"choices"`
}

//encore:api public method=POST raw path=/bootlr-chat
func BootlrChat(write http.ResponseWriter, req *http.Request){
	chatHistory, err := RetreiveChatRequestBody(write, req)
	if err != nil {
		http.Error(write, "Error retreiving request body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rlog.Info("BOOTLR CHAT", "REQ Body (chatHistory) ==> ", chatHistory)

	aiResponse, err := getAiChatResponse(chatHistory)
	if err != nil {
		http.Error(write, "Error answering the chat messsage: "+err.Error(), http.StatusInternalServerError)
		return
	}

	write.Header().Set("Content-Type", "application/json")
	response := aiResponse
	json.NewEncoder(write).Encode(response)
}
