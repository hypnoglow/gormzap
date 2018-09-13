// Package gormzap provides gorm logger implementation using Uber's zap logger.
//
// Example usage:
//  orm, _ := gorm.Open("postgres", dsn)
//  orm.LogMode(true)
//  orm.SetLogger(gormzap.New(log, gormzap.WithLevel(zap.InfoLevel))
package gormzap

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

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
// It affects only general logs, e.g. those that contain SQL queries.
// Errors will be logged with error level independently of this option.
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
		origin:      origin,
		level:       zap.DebugLevel,
		encoderFunc: DefaultRecordToFields,
	}

	for _, o := range opts {
		o(l)
	}

	return l
}

// Print implements gorm's logger interface.
func (l *Logger) Print(values ...interface{}) {
	rec := l.newRecord(values...)
	l.origin.Check(rec.Level, rec.Message).Write(l.encoderFunc(rec)...)
}

func (l *Logger) newRecord(values ...interface{}) Record {
	// See https://github.com/jinzhu/gorm/blob/master/main.go#L774
	// for info how gorm logs messages.

	if len(values) < 2 {
		// Should this ever happen?
		return Record{
			Message: fmt.Sprint(values...),
			Level:   l.level,
		}
	}

	// Handle https://github.com/jinzhu/gorm/blob/32455088f24d6b1e9a502fb8e40fdc16139dbea8/main.go#L716
	if len(values) == 2 {
		return Record{
			Message: fmt.Sprintf("%v", values[1]),
			Source:  fmt.Sprintf("%v", values[0]),
			Level:   zapcore.ErrorLevel,
		}
	}

	level := values[0]

	// Handle https://github.com/jinzhu/gorm/blob/32455088f24d6b1e9a502fb8e40fdc16139dbea8/main.go#L778
	if level == "log" {
		// By default, assume this is a user log.
		// See: https://github.com/jinzhu/gorm/blob/32455088f24d6b1e9a502fb8e40fdc16139dbea8/scope.go#L96
		// If this is an error log, we set level to error.
		// See: https://github.com/jinzhu/gorm/blob/32455088f24d6b1e9a502fb8e40fdc16139dbea8/main.go#L718
		logLevel := l.level
		if _, ok := values[2].(error); ok {
			logLevel = zapcore.ErrorLevel
		}

		return Record{
			Message: fmt.Sprint(values[2:]...),
			Source:  fmt.Sprintf("%v", values[1]),
			Level:   logLevel,
		}
	}

	// Handle https://github.com/jinzhu/gorm/blob/32455088f24d6b1e9a502fb8e40fdc16139dbea8/main.go#L786
	if level == "sql" {
		return Record{
			Message:      "gorm query",
			Source:       fmt.Sprintf("%v", values[1]),
			Duration:     values[2].(time.Duration),
			SQL:          formatSQL(values[3].(string), values[4].([]interface{})),
			RowsAffected: values[5].(int64),
			Level:        l.level,
		}
	}

	// Should this ever happen?
	return Record{
		Message: fmt.Sprint(values[2:]...),
		Source:  fmt.Sprintf("%v", values[1]),
		Level:   l.level,
	}
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
		replacements[(size-i-1)*2] = indexFunc(i)
		replacements[(size-i-1)*2+1] = formatValue(values[i])
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

func formatValue(value interface{}) string {
	indirectValue := reflect.Indirect(reflect.ValueOf(value))
	if !indirectValue.IsValid() {
		return "NULL"
	}

	value = indirectValue.Interface()

	switch v := value.(type) {
	case time.Time:
		return fmt.Sprintf("'%v'", v.Format("2006-01-02 15:04:05"))
	case []byte:
		s := string(v)
		if isPrintable(s) {
			return redactLong(fmt.Sprintf("'%s'", s))
		}
		return "'<binary>'"
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case driver.Valuer:
		if dv, err := v.Value(); err == nil && dv != nil {
			return formatValue(dv)
		}
		return "NULL"
	default:
		return redactLong(fmt.Sprintf("'%v'", value))
	}
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func redactLong(s string) string {
	if len(s) > maxLen {
		return "'<redacted>'"
	}
	return s
}

const maxLen = 255
