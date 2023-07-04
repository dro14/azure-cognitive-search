package telegram

import (
	"context"
	"github.com/dro14/profi-bot/client/azure"
	"log"

	"github.com/dro14/profi-bot/client/openai"
	"github.com/dro14/profi-bot/client/telegram"
	"github.com/dro14/profi-bot/postgres"
	"github.com/gin-gonic/gin"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Init() {
	telegram.Init()
	openai.Init()
	postgres.Init()
	azure.Init()
}

func ProcessUpdate(c *gin.Context) {

	update := &tgbotapi.Update{}
	if err := c.ShouldBindJSON(update); err != nil {
		log.Printf("can't bind json: %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	switch {
	case update.Message != nil:
		go ProcessMessage(context.Background(), update.Message)
	default:
		log.Printf("unknown update type:\n%+v", update)
	}

	c.JSON(200, gin.H{"ok": true})
}

func ProcessMessage(ctx context.Context, message *tgbotapi.Message) {

	ctx, allow := messageUpdate(ctx, message)
	if !allow {
		return
	}
	defer blockedUsers.Delete(message.From.ID)

	log.Printf("message %q from %d", message.Text, message.From.ID)
	Stream(ctx, message)
}
