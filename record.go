package gormzap

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Record is gormzap log record.
type Record struct {
	Message string
	Source  string
	Level   zapcore.Level

	Duration     time.Duration
	SQL          string
	RowsAffected int64
}

// RecordToFields func can encode gormzap Record into a slice of zap fields.
type RecordToFields func(r Record) []zapcore.Field

// DefaultRecordToFields is default encoder func for gormzap log records.
func DefaultRecordToFields(r Record) []zapcore.Field {
	// Note that Level field is ignored here, because it is handled outside
	// by zap itself.

	if r.SQL != "" {
		return []zapcore.Field{
			zap.String("sql.source", r.Source),
			zap.Duration("sql.duration", r.Duration),
			zap.String("sql.query", r.SQL),
			zap.Int64("sql.rows_affected", r.RowsAffected),
		}
	}

	return []zapcore.Field{zap.String("sql.source", r.Source)}
}
