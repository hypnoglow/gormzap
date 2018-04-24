package gormzap_test

import (
	"time"

	"go.uber.org/zap"

	"github.com/hypnoglow/gormzap"
	"go.uber.org/zap/zapcore"
)

func ExampleLogger() {
	z := zap.NewExample()

	l := gormzap.New(z)

	l.Print(
		"sql",
		"/foo/bar.go",
		time.Second * 2,
		"SELECT * FROM foo WHERE id = ?",
		[]interface{}{123},
		int64(2),
	)

	// Output:
	// {"level":"debug","msg":"gorm query","sql.source":"/foo/bar.go","sql.duration":"2s","sql.query":"SELECT * FROM foo WHERE id = 123","sql.rows_affected":2}
}

func ExampleWithRecordToFields() {
	z := zap.NewExample()

	l := gormzap.New(
		z,
		gormzap.WithLevel(zap.DebugLevel),
		gormzap.WithRecordToFields(func(r gormzap.Record) []zapcore.Field {
			return []zapcore.Field{
				zap.String("caller", r.Source),
				zap.Float32("duration_ms", float32(r.Duration.Nanoseconds() / 1000) / 1000),
				zap.String("query", r.SQL),
				zap.Int64("rows_affected", r.RowsAffected),
			}
		}),
	)

	l.Print(
		"sql",
		"/foo/bar.go",
		time.Millisecond * 200,
		"SELECT * FROM foo WHERE id = ?",
		[]interface{}{123},
		int64(2),
	)

	// Output:
	// {"level":"debug","msg":"gorm query","caller":"/foo/bar.go","duration_ms":200,"query":"SELECT * FROM foo WHERE id = 123","rows_affected":2}
}
