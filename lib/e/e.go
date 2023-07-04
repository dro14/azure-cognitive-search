package e

import "errors"

// Telegram errors
const (
	BrokenPipe         = "broken pipe"
	TooManyRequests    = "420: FLOOD_WAIT "
	MessageNotFound    = "400: MESSAGE_ID_INVALID"
	MessageEmpty       = "400: MESSAGE_EMPTY"
	MessageNotModified = "400: MESSAGE_NOT_MODIFIED"
	MessageTooLong     = "400: MESSAGE_TOO_LONG"
	UserBlocked        = "400: USER_IS_BLOCKED"
)

var (
	UserDeletedMessage = errors.New(MessageNotFound)
	UserBlockedError   = errors.New(UserBlocked)
)

// OpenAI errors
const (
	ContextLengthExceededGPT3 = "This model's maximum context length is 4097 tokens. However, you requested "
	ContextLengthExceededGPT4 = "This model's maximum context length is 8192 tokens. However, you requested "
	InvalidRequest            = "invalid_request_error"
	StreamError               = "stream error"
	BadGateway                = "502 Bad Gateway"
)

// Postgres errors
const (
	UniqueConstraint      = "violates unique constraint"
	ForeignKeyConstraint  = "violates foreign key constraint"
	UnsupportedConversion = "converting NULL to string is unsupported"
	NotFound              = "no rows in result set"
)
