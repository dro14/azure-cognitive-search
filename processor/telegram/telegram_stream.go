package telegram

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/dro14/profi-bot/client/telegram"
	"github.com/dro14/profi-bot/lib/constants"
	"github.com/dro14/profi-bot/lib/e"
	"github.com/dro14/profi-bot/lib/types"
	"github.com/dro14/profi-bot/processor/openai"
	"github.com/dro14/profi-bot/text"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Stream(ctx context.Context, message *tgbotapi.Message) {

	stats := &types.Stats{}
	stats.Requests++
	messageID, err := telegram.SendMessage(ctx, text.Loading[lang(ctx)], message.MessageID, nil)
	if err != nil {
		log.Printf("can't send loading message")
		return
	}
	beginning := ctx.Value("beginning").(time.Time)
	stats.FirstSend = time.Since(beginning).Milliseconds()
	stats.Activity = int(activity.Add(1))
	defer activity.Add(-1)

	messages := loadContext(ctx, message.Text)
	channel := make(chan string)
	go openai.ProcessWithStream(ctx, messages, stats, channel)

	isTyping := &atomic.Bool{}
	isTyping.Store(true)
	go telegram.SetTyping(ctx, isTyping)
	defer isTyping.Store(false)

	index := 0
	completion := ""
	var completions []string
	for completion = range channel {

		completions = slice(completion)
		if index >= len(completions) {
			index = len(completions) - 1
		}

		stats.Requests++
		err = telegram.EditMessage(ctx, completions[index], messageID, nil)
		if err == e.UserBlockedError {
			return
		} else if err == e.UserDeletedMessage {
			log.Printf("user deleted completion")
			index--
		}

		switch completion {
		case text.TooLong[lang(ctx)]:
			log.Printf("prompt was too long")
			fallthrough
		case text.RequestFailed[lang(ctx)]:
			return
		case text.Error[lang(ctx)]:
			index--
		}

		for index < len(completions)-1 {
			index = len(completions) - 1
			stats.Requests++
			time.Sleep(constants.RequestInterval)
			messageID, err = telegram.SendMessage(ctx, completions[index], 0, nil)
			if err == e.UserBlockedError {
				return
			} else if err != nil {
				log.Printf("can't send next message")
				index--
			}
		}

		time.Sleep(constants.RequestInterval)
	}
	stats.LastEdit = time.Since(beginning).Milliseconds()
	stats.CompletedAt = time.Now().Unix()

	storeContext(ctx, messages[1:], completion)
	//postgres.SaveMessage(ctx, stats, message.From)
}
