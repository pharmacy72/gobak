package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(conf *config.Config) (*zap.Logger, error) {
	level := zap.NewAtomicLevel()
	err := level.UnmarshalText([]byte(conf.LogLevel))
	if err != nil {
		return nil, err
	}
	logEncoding := "json"
	logEncodeLevel := zapcore.LowercaseLevelEncoder

	if conf.DevMode {
		logEncoding = "console"
		logEncodeLevel = zapcore.LowercaseColorLevelEncoder
	}

	logConfig := zap.Config{
		Level:       level,
		Development: conf.DevMode,
		Encoding:    logEncoding,
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "msg",
			LevelKey:       "severity",
			TimeKey:        "timestamp",
			EncodeLevel:    logEncodeLevel,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder,
		},
		OutputPaths: []string{"stdout"},
	}
	log, err := logConfig.Build()
	if err != nil {
		return nil, err
	}
	log = log.With(getFields(conf)...)

	return log, nil
}

func getFields(conf *config.Config) []zapcore.Field {
	var fields []zapcore.Field

	if conf.Version != "" {
		fields = append(fields, zap.String("component_version", conf.Version))
	}

	if conf.DockerId != "" {
		fields = append(fields, zap.String("docker_id", conf.DockerId))
	}

	if conf.ClsToken != "" {
		fields = append(fields, zap.String("cls_token", conf.ClsToken))
	}

	if conf.Index != "" {
		fields = append(fields, zap.String("index", conf.Index))
	}

	return fields
}
