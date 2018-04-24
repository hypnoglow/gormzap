// Package gormzap provides gorm logger implementation using Uber's zap logger.
//
// Example usage:
//  orm, _ := gorm.Open("postgres", dsn)
//  orm.LogMode(true)
//  orm.SetLogger(gormzap.New(log, gormzap.WithLevel(zap.InfoLevel))
package gormzap

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a gorm logger implementation using zap.
type Logger struct {
	origin      *zap.Logger
	level       zapcore.Level
	encoderFunc RecordToFields
}

// LoggerOption is an option for Logger.
type LoggerOption func(*Logger)

// WithLevel returns Logger option that sets level for gorm logs.
func WithLevel(level zapcore.Level) LoggerOption {
	return func(l *Logger) {
		l.level = level
	}
}

// WithRecordToFields returns Logger option that sets RecordToFields func which
// encodes log Record to a slice of zap fields.
//
// This can be used to control field names or field values types.
func WithRecordToFields(f RecordToFields) LoggerOption {
	return func(l *Logger) {
		l.encoderFunc = f
	}
}

// New returns a new gorm logger implemented using zap.
// By default it logs with debug level.
func New(origin *zap.Logger, opts ...LoggerOption) *Logger {
	l := &Logger{
		origin:  origin,
		level:   zap.DebugLevel,
		encoderFunc: DefaultRecordToFields,
	}

	for _, o := range opts {
		o(l)
	}

	return l
}

// Print implements gorm's logger interface.
func (l *Logger) Print(values ...interface{}) {
	rec := newRecord(values...)
	l.origin.Check(l.level, rec.Message).Write(l.encoderFunc(rec)...)
}

func newRecord(values ...interface{}) Record {
	var rec Record
	rec.Message = "gorm query"

	if len(values) < 1 {
		return rec
	}

	rec.Source = fmt.Sprintf("%v", values[1])

	level := values[0]
	switch level {
	case "sql":
		rec.Duration = values[2].(time.Duration)
		rec.SQL = formatSQL(values[3].(string), values[4].([]interface{}))
		rec.RowsAffected = values[5].(int64)
	default:
		rec.Message = fmt.Sprint(values[2:]...)
	}

	return rec
}

func formatSQL(sql string, values []interface{}) string {
	size := len(values)

	replacements := make([]string, size*2)

	var indexFunc func(int) string
	if strings.Contains(sql, "$1") {
		indexFunc = formatNumbered
	} else {
		indexFunc = formatQuestioned
	}

	for i := size - 1; i >= 0; i-- {
		// TODO: implement proper formatting for specific types.
		var s string
		switch values[i].(type) {
		default:
			s = fmt.Sprintf("%v", values[i])
		}

		replacements[(size-i-1)*2] = indexFunc(i)
		replacements[(size-i-1)*2+1] = s
	}

	r := strings.NewReplacer(replacements...)
	return r.Replace(sql)
}

func formatNumbered(index int) string {
	return fmt.Sprintf("$%d", index+1)
}

func formatQuestioned(index int) string {
	return "?"
}
