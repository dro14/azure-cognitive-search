package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dro14/profi-bot/lib/constants"
	"github.com/dro14/profi-bot/lib/types"
	"github.com/dro14/profi-bot/processor/openai"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Request struct {
	Search        string `json:"search"`
	QueryLanguage string `json:"queryLanguage"`
	QueryType     string `json:"queryType"`
}

type Response struct {
	Value []struct {
		Content string `json:"content"`
	} `json:"value"`
}

func Init() {

	adminKey, ok := os.LookupEnv("ADMIN_KEY")
	if !ok {
		log.Fatalf("admin key is not specified")
	}
	constants.AdminKey = adminKey
}

func Search(ctx context.Context, messages []types.Message) string {

	searchQuery := url.QueryEscape(openai.Process(ctx, messages))
	endpoint := fmt.Sprintf("https://gptkb-wuv5q7qffa7oi.search.windows.net/indexes/gptkbindex/docs?api-version=2021-04-30-Preview&search=%s&queryLanguage=ru-RU&queryType=semantic&captions=extractive&answers=extractive%%7Ccount-3&semanticConfiguration=default", searchQuery)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		log.Printf("can't create request: %v", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", constants.AdminKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("can't do request: %v", err)
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	response := &Response{}
	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("can't read response: %v", err)
		return ""
	}

	err = json.Unmarshal(bts, response)
	if err != nil {
		log.Printf("can't decode response: %v", err)
		return ""
	}

	var searchResult string
	for i := 0; i < 3; i++ {
		searchResult += fmt.Sprintf("\"\"\"\n%s\n\"\"\"\n\n", response.Value[i].Content)
	}

	return fmt.Sprintf(messageTemplate, searchResult)
}

var messageTemplate = `Отвечай ТОЛЬКО фактами, указанными в списке источников ниже. Если ниже недостаточно информации, скажи, что не знаешь. Не создавай ответы, в которых не используются приведенные ниже источники. НЕ ОТВЕЧАЙ на темы отличные от тем источников. Если тебе поможет уточняющий вопрос, задай вопрос.

Источники:

%s`
