package telegram

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dro14/profi-bot/lib/constants"
	"github.com/dro14/profi-bot/lib/e"
	"github.com/dro14/profi-bot/lib/functions"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

var bot *tg.Client

func Init() {

	token, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		log.Fatalf("bot token is not specified")
	}

	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		log.Fatalf("can't initialize client: %v", err)
	}

	done := make(chan bool)
	go func() {
		if err = client.Run(context.Background(), func(ctx context.Context) error {
			_, err = client.Auth().Bot(ctx, token)
			if err != nil {
				log.Fatalf("can't authorize as bot: %v", err)
			}
			bot = client.API()
			done <- true
			return telegram.RunUntilCanceled(ctx, client)
		}); err != nil {
			log.Fatalf("can't connect client: %v", err)
		}
	}()
	<-done
}

func SendMessage(ctx context.Context, message string, replyToMsgID int, keyboard *tg.ReplyInlineMarkup) (int, error) {

	userID := ctx.Value("user_id").(int64)

	request := &tg.MessagesSendMessageRequest{
		Peer:         &tg.InputPeerUser{UserID: userID},
		Message:      message,
		ReplyToMsgID: replyToMsgID,
		RandomID:     time.Now().UnixNano(),
		NoWebpage:    true,
	}

	if keyboard != nil {
		request.ReplyMarkup = keyboard
	}

	retryDelay := constants.RetryDelay
	attempts := 0
Retry:
	attempts++
	resp, err := bot.MessagesSendMessage(ctx, request)
	if err != nil {

		log.Printf("can't send message to %d: %v", userID, err)

		switch {
		case strings.Contains(err.Error(), e.BrokenPipe):
			log.Fatalf("fatal error: restarting bot")
		case strings.Contains(err.Error(), e.UserBlocked):
			return 0, e.UserBlockedError
		case strings.Contains(err.Error(), e.MessageEmpty),
			strings.Contains(err.Error(), e.MessageTooLong):
			log.Printf("%q", request.Message)
			return 0, err
		case strings.Contains(err.Error(), e.TooManyRequests):
			_, str, _ := strings.Cut(err.Error(), e.TooManyRequests)
			str = str[1 : len(str)-1]
			seconds, _ := strconv.Atoi(str)
			retryDelay = time.Duration(seconds) * time.Second
		}

		if attempts < constants.RetryAttempts {
			functions.Sleep(&retryDelay)
			goto Retry
		}
		return 0, err
	}

	response, ok := resp.(*tg.UpdateShortSentMessage)
	if !ok {
		log.Printf("can't decode response for %d: %v", userID, err)
		return 0, fmt.Errorf("can't decode response for %d", userID)
	}

	if attempts > 1 {
		log.Printf("sending message to %d was handled after %d attempts", userID, attempts)
	}

	return response.ID, nil
}

func EditMessage(ctx context.Context, message string, messageID int, keyboard *tg.ReplyInlineMarkup) error {

	userID := ctx.Value("user_id").(int64)

	request := &tg.MessagesEditMessageRequest{
		Peer:      &tg.InputPeerUser{UserID: userID},
		Message:   message,
		ID:        messageID,
		NoWebpage: true,
	}

	if keyboard != nil {
		request.ReplyMarkup = keyboard
	}

	retryDelay := constants.RetryDelay
	attempts := 0
Retry:
	attempts++
	_, err := bot.MessagesEditMessage(ctx, request)
	if err != nil {

		log.Printf("can't edit message for %d: %v", userID, err)

		switch {
		case strings.Contains(err.Error(), e.BrokenPipe):
			log.Fatalf("fatal error: restarting bot")
		case strings.Contains(err.Error(), e.UserBlocked):
			return e.UserBlockedError
		case strings.Contains(err.Error(), e.MessageNotFound):
			return e.UserDeletedMessage
		case strings.Contains(err.Error(), e.MessageEmpty),
			strings.Contains(err.Error(), e.MessageTooLong),
			strings.Contains(err.Error(), e.MessageNotModified):
			log.Printf("%q", request.Message)
			return err
		case strings.Contains(err.Error(), e.TooManyRequests):
			_, str, _ := strings.Cut(err.Error(), e.TooManyRequests)
			str = str[1 : len(str)-1]
			seconds, _ := strconv.Atoi(str)
			retryDelay = time.Duration(seconds) * time.Second
		}

		if attempts < constants.RetryAttempts {
			functions.Sleep(&retryDelay)
			goto Retry
		}
		return err
	}

	if attempts > 1 {
		log.Printf("editing message for %d was handled after %d attempts", userID, attempts)
	}

	return nil
}

func SetTyping(ctx context.Context, isTyping *atomic.Bool) {

	userID := ctx.Value("user_id").(int64)

	request := &tg.MessagesSetTypingRequest{
		Peer:   &tg.InputPeerUser{UserID: userID},
		Action: &tg.SendMessageTypingAction{},
	}

Loop:
	for isTyping.Load() {
		_, err := bot.MessagesSetTyping(ctx, request)
		if err != nil {

			log.Printf("can't set typing for %d: %v", userID, err)

			switch {
			case strings.Contains(err.Error(), e.BrokenPipe):
				log.Fatalf("fatal error: restarting bot")
			case strings.Contains(err.Error(), e.UserBlocked):
				break Loop
			}
		}
		time.Sleep(5800 * time.Millisecond)
	}
}
