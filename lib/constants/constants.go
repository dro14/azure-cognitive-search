package constants

import "time"

const (
	RetryAttempts = 10
)

const (
	RetryDelay      = 1000 * time.Millisecond
	RequestInterval = 1500 * time.Millisecond
)

var (
	MerchantID      string
	TestKey         string
	RealKey         string
	SubscriptionKey string
	AdminKey        string
)
