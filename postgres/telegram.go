package postgres

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/dro14/profi-bot/lib/constants"
	"github.com/dro14/profi-bot/lib/e"
	"github.com/dro14/profi-bot/lib/functions"
	"github.com/dro14/profi-bot/lib/types"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func JoinUser(ctx context.Context, user *tgbotapi.User, joinedBy int64) {

	var (
		attempts   int
		retryDelay = constants.RetryDelay
		lang       = ctx.Value("language_code").(string)
		date       = int64(ctx.Value("date").(int))
		joinedAt   = time.Unix(date, 0).Format(time.DateTime)
	)

Retry:
	attempts++
	_, _ = db.ExecContext(ctx, `
		INSERT INTO users
    		(id, first_name, last_name, username, language_code)
		VALUES
			($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING;`,
		user.ID, user.FirstName, user.LastName, user.UserName, lang)

	_, err := db.ExecContext(ctx, `
		INSERT INTO user_configs
			(id, joined_by, is_active, joined_at)
		VALUES
		    ($1, $2, $3, $4);`,
		user.ID, joinedBy, true, joinedAt)

	if err != nil {
		if strings.Contains(err.Error(), e.UniqueConstraint) {
			if !IsActive(ctx, user) {
				RejoinUser(ctx, user)
			}
		} else {
			log.Printf("can't join user: %v", err)
			if attempts < constants.RetryAttempts {
				functions.Sleep(&retryDelay)
				goto Retry
			} else {
				log.Printf("joining user failed after %d attempts", attempts)
			}
		}
	}
}

func SaveMessage(ctx context.Context, stats *types.Stats, user *tgbotapi.User) {

	var (
		attempts    int
		retryDelay  = constants.RetryDelay
		lang        = ctx.Value("language_code").(string)
		date        = int64(ctx.Value("date").(int))
		createdOn   = time.Unix(date, 0).Format(time.DateOnly)
		promptedAt  = time.Unix(date, 0).Format(time.TimeOnly)
		completedAt = time.Unix(stats.CompletedAt, 0).Format(time.TimeOnly)
	)

	if !IsActive(ctx, user) {
		RejoinUser(ctx, user)
	}

Retry:
	attempts++
	_, err := db.ExecContext(ctx, `
		INSERT INTO messages
			(user_id, is_premium, created_on, prompted_at, completed_at, first_send, last_edit, prompt_tokens, prompt_length, completion_tokens, completion_length, activity, requests, attempts, finish_reason, language_code)
		VAlUES
		    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16);`,
		user.ID, createdOn, promptedAt, completedAt, stats.FirstSend, stats.LastEdit, stats.PromptTokens, stats.PromptLength, stats.CompletionTokens, stats.CompletionLength, stats.Activity, stats.Requests, stats.Attempts, stats.FinishReason, lang)

	if err != nil {
		if strings.Contains(err.Error(), e.ForeignKeyConstraint) {
			JoinUser(ctx, user, 0)
			goto Retry
		} else {
			log.Printf("can't save message: %v", err)
			if attempts < constants.RetryAttempts {
				functions.Sleep(&retryDelay)
				goto Retry
			} else {
				log.Printf("saving message failed after %d attempts", attempts)
			}
		}
	}
}

func DeactivateUser(ctx context.Context, user *tgbotapi.User) {

	var (
		attempts      int
		retryDelay    = constants.RetryDelay
		date          = int64(ctx.Value("date").(int))
		deactivatedAt = time.Unix(date, 0).Format(time.DateTime)
	)

Retry:
	attempts++
	_, err := db.ExecContext(ctx, `
	UPDATE user_configs
	SET
		is_active = $1,
	    deactivated_at = $2
	WHERE
	    id = $3;`,
		false, deactivatedAt, user.ID)

	if err != nil {
		if strings.Contains(err.Error(), e.NotFound) {
			JoinUser(ctx, user, 0)
			goto Retry
		} else {
			log.Printf("can't deactivate user: %v", err)
			if attempts < constants.RetryAttempts {
				functions.Sleep(&retryDelay)
				goto Retry
			} else {
				log.Printf("deactivating user failed after %d attempts", attempts)
			}
		}
	}
}

func RejoinUser(ctx context.Context, user *tgbotapi.User) {

	var (
		attempts   int
		retryDelay = constants.RetryDelay
		date       = int64(ctx.Value("date").(int))
		rejoinedAt = time.Unix(date, 0).Format(time.DateTime)
	)

Retry:
	attempts++
	_, err := db.ExecContext(ctx, `
	UPDATE user_configs
	SET
	    is_active = $1,
	    rejoined_at = $2
	WHERE
	    id = $3;`,
		true, rejoinedAt, user.ID)

	if err != nil {
		if strings.Contains(err.Error(), e.NotFound) {
			JoinUser(ctx, user, 0)
			goto Retry
		} else {
			log.Printf("can't rejoin user: %v", err)
			if attempts < constants.RetryAttempts {
				functions.Sleep(&retryDelay)
				goto Retry
			} else {
				log.Printf("rejoining user failed after %d attempts", attempts)
			}
		}
	}
}

func IsActive(ctx context.Context, user *tgbotapi.User) bool {

	var (
		attempts   int
		retryDelay = constants.RetryDelay
		isActive   bool
	)

Retry:
	attempts++
	err := db.QueryRowContext(ctx, `
		SELECT is_active
		FROM user_configs
		WHERE id = $1;`,
		user.ID).Scan(&isActive)

	if err != nil {
		if strings.Contains(err.Error(), e.NotFound) {
			JoinUser(ctx, user, 0)
			goto Retry
		} else {
			log.Printf("can't get is_active: %v", err)
			if attempts < constants.RetryAttempts {
				functions.Sleep(&retryDelay)
				goto Retry
			} else {
				log.Printf("getting is_active failed after %d attempts", attempts)
			}
		}
	}

	return isActive
}

func JoinedAt(ctx context.Context, user *tgbotapi.User) string {

	var (
		attempts   int
		retryDelay = constants.RetryDelay
		joinedAt   string
	)

Retry:
	attempts++
	err := db.QueryRowContext(ctx, `
		SELECT joined_at
		FROM user_configs
		WHERE id = $1;`,
		user.ID).Scan(&joinedAt)

	if err != nil {
		if strings.Contains(err.Error(), e.NotFound) {
			JoinUser(ctx, user, 0)
			goto Retry
		} else {
			log.Printf("can't get joined_at: %v", err)
			if attempts < constants.RetryAttempts {
				functions.Sleep(&retryDelay)
				goto Retry
			} else {
				log.Printf("getting joined_at failed after %d attempts", attempts)
			}
		}
	}

	return joinedAt
}

func DeactivatedAt(ctx context.Context, user *tgbotapi.User) string {

	var (
		attempts      int
		retryDelay    = constants.RetryDelay
		deactivatedAt string
	)

Retry:
	attempts++
	err := db.QueryRowContext(ctx, `
		SELECT deactivated_at
		FROM user_configs
		WHERE id = $1;`,
		user.ID).Scan(&deactivatedAt)

	if err != nil {
		if strings.Contains(err.Error(), e.UnsupportedConversion) {
			return ""
		} else if strings.Contains(err.Error(), e.NotFound) {
			JoinUser(ctx, user, 0)
			goto Retry
		} else {
			log.Printf("can't get deactivated_at: %v", err)
			if attempts < constants.RetryAttempts {
				functions.Sleep(&retryDelay)
				goto Retry
			} else {
				log.Printf("getting deactivated_at failed after %d attempts", attempts)
			}
		}
	}

	return deactivatedAt
}

func RejoinedAt(ctx context.Context, user *tgbotapi.User) string {

	var (
		attempts   int
		retryDelay = constants.RetryDelay
		rejoinedAt string
	)

Retry:
	attempts++
	err := db.QueryRowContext(ctx, `
		SELECT rejoined_at
		FROM user_configs
		WHERE id = $1;`,
		user.ID).Scan(&rejoinedAt)

	if err != nil {
		if strings.Contains(err.Error(), e.UnsupportedConversion) {
			return ""
		} else if strings.Contains(err.Error(), e.NotFound) {
			JoinUser(ctx, user, 0)
			goto Retry
		} else {
			log.Printf("can't get rejoined_at: %v", err)
			if attempts < constants.RetryAttempts {
				functions.Sleep(&retryDelay)
				goto Retry
			} else {
				log.Printf("getting rejoined_at failed after %d attempts", attempts)
			}
		}
	}

	return rejoinedAt
}
