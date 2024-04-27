package bootlrChat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"encore.app/utils"
)

type ChatResponse struct {
	ResponseText     string   `json:"responseText"`
	ProductReference []string `json:"productReference"`
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


//encore:api public method=POST raw path=/bootlr-chat
func BootlrChat(write http.ResponseWriter, req *http.Request){
	chatHistory, err := RetreiveChatRequestBody(write, req)
	if err != nil {
		http.Error(write, "Error retreiving request body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	aiResponse, err := getAiChatResponse(chatHistory)
	if err != nil {
		http.Error(write, "Error answering the chat messsage: "+err.Error(), http.StatusInternalServerError)
		return
	}

	write.Header().Set("Content-Type", "application/json")
	response := aiResponse
	json.NewEncoder(write).Encode(response)
}

// =================
// HELPER FUNCTIONS
// =================

func RetreiveChatRequestBody(write http.ResponseWriter, req *http.Request) ([]MessageHistoryItem, error) {
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

func getAiChatResponse(messageHistory []MessageHistoryItem) (string, error) {
	OPENAI_KEY := utils.Secrets.OPENAI_KEY
	OPENAI_ORG := utils.Secrets.OPENAI_ORG
	OPENAI_URL := "https://api.openai.com/v1/chat/completions"

	openAiRequestBody := OpenAIRequest{
			Model:       "gpt-3.5-turbo",
			Messages:    messageHistory,
			Temperature: 0.7,
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
	
	chatResponse :=  openAIResponse.Choices[0].Message.Content
	return chatResponse, nil
}

