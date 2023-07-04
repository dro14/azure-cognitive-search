package types

type Stats struct {
	CompletedAt                        int64
	FirstSend, LastEdit                int64
	Prompt, Completion                 string
	PromptTokens, PromptLength         int
	CompletionTokens, CompletionLength int
	Activity, Requests, Attempts       int
	FinishReason                       string
}
