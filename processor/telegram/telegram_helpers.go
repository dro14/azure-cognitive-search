package telegram

import (
	"context"
	"github.com/dro14/profi-bot/client/azure"
	"github.com/dro14/profi-bot/lib/types"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dro14/profi-bot/lib/functions"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	blockedUsers sync.Map
	Context      sync.Map
	activity     atomic.Int64
)

func isBlocked(userID int64) bool {

	_, ok := blockedUsers.Load(userID)
	if ok {
		return true
	}

	blockedUsers.Store(userID, true)
	go unblockUser(userID)
	return false
}

func unblockUser(userID int64) {
	time.Sleep(1 * time.Minute)
	blockedUsers.Delete(userID)
}

func messageUpdate(ctx context.Context, message *tgbotapi.Message) (context.Context, bool) {

	switch {
	case message.From.IsBot,
		len(message.Text) == 0,
		message.Chat.Type != "private",
		isBlocked(message.From.ID):
		return ctx, false
	}

	ctx = context.WithValue(ctx, "beginning", time.Now())
	ctx = context.WithValue(ctx, "date", message.Date)
	ctx = context.WithValue(ctx, "user_id", message.From.ID)
	ctx = context.WithValue(ctx, "language_code", functions.LanguageCode(message.From.LanguageCode))
	ctx = context.WithValue(ctx, "model", "gpt-3.5-turbo")
	return ctx, true
}

func lang(ctx context.Context) string {
	return ctx.Value("language_code").(string)
}

func slice(completion string) []string {

	var completions []string

	for len(completion) > 4096 {
		cutIndex := 0
	Loop:
		for i := 4096; i >= 0; i-- {
			switch completion[i] {
			case ' ', '\n', '\t', '\r':
				cutIndex = i
				break Loop
			}
		}
		completions = append(completions, completion[:cutIndex])
		completion = completion[cutIndex:]
	}

	return append(completions, completion)
}

func loadContext(ctx context.Context, prompt string) []types.Message {

	userID := ctx.Value("user_id").(int64)

	var messages []types.Message
	value, ok := Context.Load(userID)
	if ok {
		for _, message := range value.([]types.Message) {
			messages = append(messages, message)
		}
	}
	messages = append(messages, types.Message{Role: "user", Content: prompt})

	systemMessage := azure.Search(ctx, messages)
	if len(systemMessage) > 0 {
		messages = append([]types.Message{{
			Role:    "system",
			Content: systemMessage,
		}}, messages...)
	}

	return messages
}

func storeContext(ctx context.Context, messages []types.Message, completion string) {

	userID := ctx.Value("user_id").(int64)

	messages = append(messages, types.Message{Role: "assistant", Content: completion})
	if len(messages) > 6 {
		messages = messages[2:]
	}
	Context.Store(userID, messages)
}
