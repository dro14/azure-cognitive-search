package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"github.com/dro14/profi-bot/lib/types"
)

var keys []string
var index int

func Init() {

	for i := 0; ; i++ {
		key := fmt.Sprintf("OPENAI_TOKEN_%d", i)
		token, ok := os.LookupEnv(key)
		if !ok {
			break
		}
		keys = append(keys, "Bearer "+token)
	}

	if len(keys) == 0 {
		log.Fatalf("openai token is not specified")
	}
}

func CompletionWithStream(ctx context.Context, messages []types.Message, maxTokens int, channel chan<- string) (*types.Response, error) {

	request := &types.Request{
		Model:     "gpt-3.5-turbo-0613",
		Messages:  messages,
		MaxTokens: maxTokens,
		Stream:    true,
		User:      fmt.Sprintf("%d", ctx.Value("user_id").(int64)),
	}

	resp, err := send(ctx, request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	buffer := &atomic.Value{}
	buffer.Store("")
	isStreaming := &atomic.Int64{}
	isStreaming.Store(1)
	go streamOut(buffer, isStreaming, channel)

	response, err := streamIn(resp, buffer)
	if err != nil {
		isStreaming.Store(-1)
		return nil, err
	}

	isStreaming.Store(0)
	return response, nil
}

func Completion(ctx context.Context, messages []types.Message, maxTokens int) (string, error) {

	request := &types.Request{
		Model:     "gpt-3.5-turbo-0613",
		Messages:  messages,
		MaxTokens: maxTokens,
		User:      fmt.Sprintf("%d", ctx.Value("user_id").(int64)),
	}

	resp, err := send(ctx, request)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	response := &types.Response{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Printf("can't decode response: %v", err)
		return "", err
	}

	if len(strings.TrimSpace(response.Choices[0].Message.Content)) == 0 {
		log.Printf("empty response for %d", ctx.Value("user_id").(int64))
		return "", fmt.Errorf("empty response for %d", ctx.Value("user_id").(int64))
	}

	return response.Choices[0].Message.Content, nil
}
