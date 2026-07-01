package logging

import (
	"github.com/ThreeDotsLabs/watermill"
	"go.uber.org/zap"
)

var _ watermill.LoggerAdapter = (*ZapWatermillLoggerAdapter)(nil)

type ZapWatermillLoggerAdapter struct {
	logger *zap.Logger
}

func NewZapWatermillLoggerAdapter(logger *zap.Logger) *ZapWatermillLoggerAdapter {
	return &ZapWatermillLoggerAdapter{logger: logger}
}

func zapFields(fields watermill.LogFields) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}

func (z *ZapWatermillLoggerAdapter) Error(msg string, err error, fields watermill.LogFields) {
	z.logger.Error(msg, append(zapFields(fields), zap.Error(err))...)
}

func (z *ZapWatermillLoggerAdapter) Info(msg string, fields watermill.LogFields) {
	z.logger.Info(msg, zapFields(fields)...)
}

func (z *ZapWatermillLoggerAdapter) Debug(msg string, fields watermill.LogFields) {
	z.logger.Debug(msg, zapFields(fields)...)
}

func (z *ZapWatermillLoggerAdapter) Trace(msg string, fields watermill.LogFields) {
	z.logger.Debug(msg, zapFields(fields)...)
}

func (z *ZapWatermillLoggerAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return &ZapWatermillLoggerAdapter{logger: z.logger.With(zapFields(fields)...)}
}
