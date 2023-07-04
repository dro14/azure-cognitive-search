package openai

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dro14/profi-bot/client/bobdev"
	"github.com/dro14/profi-bot/client/openai"
	"github.com/dro14/profi-bot/lib/constants"
	"github.com/dro14/profi-bot/lib/e"
	"github.com/dro14/profi-bot/lib/functions"
	"github.com/dro14/profi-bot/lib/types"
	"github.com/dro14/profi-bot/text"
)

func ProcessWithStream(ctx context.Context, messages []types.Message, stats *types.Stats, channel chan<- string) {

	maxTokens := 4096 - bobdev.Tokens(ctx, messages)
	retryDelay := 10 * constants.RetryDelay
	var errMsg string
Retry:
	stats.Attempts++
	response, err := openai.CompletionWithStream(ctx, messages, maxTokens, channel)
	if err != nil {
		errMsg = err.Error()

		switch {
		case strings.HasPrefix(errMsg, e.InvalidRequest):
			errMsg = strings.TrimPrefix(errMsg, e.InvalidRequest)
			if strings.HasPrefix(errMsg, e.ContextLengthExceededGPT3) {
				errMsg = strings.TrimPrefix(errMsg, e.ContextLengthExceededGPT3)
				errMsg, _, _ = strings.Cut(errMsg, " tokens")
				totalTokens, _ := strconv.Atoi(errMsg)
				diff := totalTokens - 4096
				maxTokens -= diff
			} else if len(messages) > 3 {
				messages = append(messages[:1], messages[3:]...)
				maxTokens = 4096 - bobdev.Tokens(ctx, messages)
			} else {
				channel <- text.TooLong[lang(ctx)]
				return
			}
			goto Retry
		case strings.HasPrefix(errMsg, e.StreamError):
			channel <- text.Error[lang(ctx)]
			goto Retry
		case strings.HasPrefix(errMsg, e.BadGateway):
			goto Retry
		}

		if stats.Attempts < constants.RetryAttempts {
			functions.Sleep(&retryDelay)
			goto Retry
		} else {
			log.Printf("%q failed after %d attempts", errMsg, stats.Attempts)
			channel <- text.RequestFailed[lang(ctx)]
			return
		}
	} else if stats.Attempts > 1 {
		log.Printf("%q was handled after %d attempts", errMsg, stats.Attempts)
	}

	stats.FinishReason = response.Choices[0].FinishReason
	stats.PromptTokens = 4096 - maxTokens
	stats.PromptLength = length(messages)

	completions := []types.Message{response.Choices[0].Message}
	stats.CompletionTokens = bobdev.Tokens(ctx, completions) - 8
	stats.CompletionLength = len(completions[0].Content)
}

func Process(ctx context.Context, messages []types.Message) string {

	historyText := getChatHistoryAsText(messages[:len(messages)-1])
	question := messages[len(messages)-1].Content

	messages = []types.Message{{
		Role:    "user",
		Content: fmt.Sprintf(promptQueryTemplate, historyText, question),
	}}

	maxTokens := 4096 - bobdev.Tokens(ctx, messages)
	retryDelay := 10 * constants.RetryDelay
	var errMsg string
	var attempts int
Retry:
	attempts++
	completion, err := openai.Completion(ctx, messages, maxTokens)
	if err != nil {
		errMsg = err.Error()

		switch {
		case strings.HasPrefix(errMsg, e.InvalidRequest):
			errMsg = strings.TrimPrefix(errMsg, e.InvalidRequest)
			if strings.HasPrefix(errMsg, e.ContextLengthExceededGPT3) {
				errMsg = strings.TrimPrefix(errMsg, e.ContextLengthExceededGPT3)
				errMsg, _, _ = strings.Cut(errMsg, " tokens")
				totalTokens, _ := strconv.Atoi(errMsg)
				diff := totalTokens - 4096
				maxTokens -= diff
			} else if len(messages) > 3 {
				messages = append(messages[:1], messages[3:]...)
				maxTokens = 4096 - bobdev.Tokens(ctx, messages)
			} else {
				return question
			}
			goto Retry
		}

		if attempts < constants.RetryAttempts {
			functions.Sleep(&retryDelay)
			goto Retry
		} else {
			log.Printf("%q failed after %d attempts", errMsg, attempts)
			return question
		}
	} else if attempts > 1 {
		log.Printf("%q was handled after %d attempts", errMsg, attempts)
	}

	return completion
}

var promptQueryTemplate = `Ниже приведена история разговора и новый вопрос, заданный пользователем, на который необходимо ответить, выполнив поиск в базе соответствующих данных.
     Создай поисковый запрос на основе беседы и нового вопроса.
     Не включай цитируемые имена исходных файлов и названия документов, например, info.txt или doc.pdf, в условия поискового запроса.
     Не включай текст поискового запроса внутри [] или <<>>.
     Если вопрос не на русском языке, переведи вопрос на русский язык перед созданием поискового запроса.

История разговора:
%s
Вопрос:
%s

Поисковый запрос:`
