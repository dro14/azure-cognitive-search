package bobdev

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/dro14/profi-bot/lib/types"
)

type Request struct {
	Model    string          `json:"model"`
	Messages []types.Message `json:"messages"`
}

func Tokens(ctx context.Context, messages []types.Message) int {

	request := &Request{
		Model:    ctx.Value("model").(string),
		Messages: messages,
	}

	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(request)
	if err != nil {
		log.Printf("can't encode request: %v", err)
		return 0
	}

	resp, err := http.Post("https://chatgpt-payment.herokuapp.com/tiktoken", "application/json", &buffer)
	if err != nil {
		log.Printf("can't send request: %v", err)
		return 0
	}

	response := make(map[string]int)
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return 0
	}
	_ = resp.Body.Close()

	return response["tokens"]
}
