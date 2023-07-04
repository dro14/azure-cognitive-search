package openai

import (
	"context"
	"fmt"

	"github.com/dro14/profi-bot/lib/types"
)

func length(messages []types.Message) int {
	var promptLength int
	for i := range messages {
		promptLength += len(fmt.Sprintf("role: %s\ncontent: %s", messages[i].Role, messages[i].Content))
	}
	return promptLength
}

func lang(ctx context.Context) string {
	return ctx.Value("language_code").(string)
}

func getChatHistoryAsText(messages []types.Message) string {

	var text string
	for i := range messages {
		text += fmt.Sprintf("<|im_start|>%s\n%s\n<|im_end|>\n", messages[i].Role, messages[i].Content)
	}
	return text
}
