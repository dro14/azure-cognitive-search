package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/dro14/profi-bot/lib/e"
	"github.com/dro14/profi-bot/lib/types"
)

func send(ctx context.Context, request *types.Request) (*http.Response, error) {

	userID := ctx.Value("user_id").(int64)

	resp, err := makeRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return resp, nil
	case http.StatusBadGateway, http.StatusServiceUnavailable:
		return nil, fmt.Errorf("%s for %d", resp.Status, userID)
	default:
		bts, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		var response types.Response
		err = json.Unmarshal(bts, &response)

		switch response.Error.Type {
		case e.InvalidRequest:
			return nil, fmt.Errorf(response.Error.Type + response.Error.Message)
		default:
			if err != nil {
				log.Printf("%s for %d\ncan't decode response: %v\nbody: %q", resp.Status, userID, err, string(bts))
			} else {
				log.Printf("%s for %d\ntype: %s\nmessage: %s", resp.Status, userID, response.Error.Type, response.Error.Message)
			}
			return nil, fmt.Errorf("%s for %d", resp.Status, userID)
		}
	}
}

func makeRequest(ctx context.Context, request *types.Request) (*http.Response, error) {

	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(request)
	if err != nil {
		log.Printf("can't encode request: %v", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", &buffer)
	if err != nil {
		log.Printf("can't create request: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", keys[index])

	index++
	if index == len(keys) {
		index = 0
	}

	var client http.Client
	client.Timeout = 10 * time.Minute

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("can't do request: %v", err)
		return nil, err
	}

	return resp, nil
}
