package logger

import (
	"github.com/getsentry/sentry-go"
	"github.com/pharmacy72/gobak/config"
)

func InitSentry(conf *config.Config) error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: conf.SentryDSN,
	})
	if err != nil {
		return err
	}
	return nil
}
