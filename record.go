package gormzap

import (
	"time"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap"
)

// Record is gormzap log record.
type Record struct {
	Message      string
	Source       string
	Duration     time.Duration
	SQL          string
	RowsAffected int64
}

// RecordFielder can encode gormzap Record into a slice of zap fields.
type RecordFielder interface {
	RecordFields(r Record) []zapcore.Field
}

// DefaultRecordFielder is default encoder for gormzap log records.
var DefaultRecordFielder RecordFielder = defaultFielder{}

type defaultFielder struct{}

// RecordFields implements RecordFielder.
func (defaultFielder) RecordFields(r Record) []zapcore.Field {
	return []zapcore.Field{
		zap.String("source", r.Source),
		zap.Duration("duration", r.Duration),
		zap.String("sql", r.SQL),
		zap.Int64("rows_affected", r.RowsAffected),
	}
}
