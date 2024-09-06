package bootlrChat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"encore.dev/rlog"
)


// =================
// CHAT SERVICES
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
	OPENAI_KEY := secrets.OPENAI_KEY
	OPENAI_ORG := secrets.OPENAI_ORG
	OPENAI_URL := "https://api.openai.com/v1/chat/completions"

	openAiRequestBody := OpenAIRequest{
			Model:       "gpt-4o-mini",
			Messages:    messageHistory,
			Temperature: 0.9,
			Response_format: OpenaiResponseformat{ Type : "json_object"},
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

	rlog.Info("BOOTLR CHAT", "OpenAI responseData ==> ", string(responseData))
	
	var openAIResponse OpenAIResponse
	if err := json.Unmarshal(responseData, &openAIResponse); err != nil {
		return "", err
	}
	
	chatResponse :=  openAIResponse.Choices[0].Message.Content
	return chatResponse, nil
}
