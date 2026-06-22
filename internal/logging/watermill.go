package logging

import (
	"github.com/ThreeDotsLabs/watermill"
	"go.uber.org/zap"
)

var _ watermill.LoggerAdapter = (*ZapLoggerAdapter)(nil)

type ZapLoggerAdapter struct {
	logger *zap.Logger
}

func NewZapLoggerAdapter(logger *zap.Logger) *ZapLoggerAdapter {
	return &ZapLoggerAdapter{logger: logger}
}

func zapFields(fields watermill.LogFields) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}

func (z *ZapLoggerAdapter) Error(msg string, err error, fields watermill.LogFields) {
	z.logger.Error(msg, append(zapFields(fields), zap.Error(err))...)
}

func (z *ZapLoggerAdapter) Info(msg string, fields watermill.LogFields) {
	z.logger.Info(msg, zapFields(fields)...)
}

func (z *ZapLoggerAdapter) Debug(msg string, fields watermill.LogFields) {
	z.logger.Debug(msg, zapFields(fields)...)
}

func (z *ZapLoggerAdapter) Trace(msg string, fields watermill.LogFields) {
	z.logger.Debug(msg, zapFields(fields)...)
}

func (z *ZapLoggerAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return &ZapLoggerAdapter{logger: z.logger.With(zapFields(fields)...)}
}
