package commonutils

import "errors"

const (
	PublishThreshold                    = 600
	RandomIntRange                      = 1_000_000
	TimeFormat                          = "20060102150405"
	MaxIndividualQueryCountThreshold    = 10
	ExplainTPS                          = 5
	DefaultStatementTimeoutMilliseconds = 5000
)

var (
	ErrUnsupportedVersion = errors.New("unsupported PostgreSQL version")
	ErrUnExpectedError    = errors.New("unexpected error")
	ErrInvalidModelType   = errors.New("invalid model type")
	ErrNotEligible        = errors.New("not eligible to fetch metrics")
)

const (
	PostgresVersion11 = 11
	PostgresVersion12 = 12
	PostgresVersion13 = 13
	PostgresVersion14 = 14
)
