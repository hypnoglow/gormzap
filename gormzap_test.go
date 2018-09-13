package gormzap_test

import (
	"errors"
	"testing"
	"time"

	"github.com/hypnoglow/gormzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func ExampleLogger() {
	z := zap.NewExample()

	l := gormzap.New(z)

	l.Print(
		"sql",
		"/foo/bar.go",
		time.Second*2,
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
				zap.Float32("duration_ms", float32(r.Duration.Nanoseconds()/1000)/1000),
				zap.String("query", r.SQL),
				zap.Int64("rows_affected", r.RowsAffected),
			}
		}),
	)

	l.Print(
		"sql",
		"/foo/bar.go",
		time.Millisecond*200,
		"SELECT * FROM foo WHERE id = ?",
		[]interface{}{123},
		int64(2),
	)

	// Output:
	// {"level":"debug","msg":"gorm query","caller":"/foo/bar.go","duration_ms":200,"query":"SELECT * FROM foo WHERE id = 123","rows_affected":2}
}

func TestLogger_Print(t *testing.T) {
	t.Run("log with values < 2", func(t *testing.T) {
		l, buf := logger()

		l.Print("idunno")
		expected := `{"level":"debug","msg":"idunno","sql.source":""}`

		actual := buf.Lines()[0]
		if actual != expected {
			t.Fatalf("Expected %s but got %s", expected, actual)
		}
	})

	t.Run("log with values = 2 (error)", func(t *testing.T) {
		l, buf := logger()

		l.Print("/some/file.go:32", errors.New("some serious error!"))
		expected := `{"level":"error","msg":"some serious error!","sql.source":"/some/file.go:32"}`

		actual := buf.Lines()[0]
		if actual != expected {
			t.Fatalf("Expected %s but got %s", expected, actual)
		}
	})

	t.Run("log with level = log (error)", func(t *testing.T) {
		l, buf := logger()

		l.Print(
			"log",
			"/some/file.go:33",
			errors.New("some serious error!"),
		)
		expected := `{"level":"error","msg":"some serious error!","sql.source":"/some/file.go:33"}`

		actual := buf.Lines()[0]
		if actual != expected {
			t.Fatalf("Expected %s but got %s", expected, actual)
		}
	})

	t.Run("log with level = log (user log)", func(t *testing.T) {
		l, buf := logger()

		l.Print(
			"log",
			"/some/file.go:33",
			"foo",
			"bar",
		)
		expected := `{"level":"debug","msg":"foobar","sql.source":"/some/file.go:33"}`

		actual := buf.Lines()[0]
		if actual != expected {
			t.Fatalf("Expected %s but got %s", expected, actual)
		}
	})

	t.Run("log with level = sql", func(t *testing.T) {
		l, buf := logger()

		l.Print(
			"sql",
			"/some/file.go:34",
			time.Millisecond*5,
			"SELECT * FROM test WHERE id = $1",
			[]interface{}{42},
			int64(1),
		)
		expected := `{"level":"debug","msg":"gorm query","sql.source":"/some/file.go:34","sql.duration":"5ms","sql.query":"SELECT * FROM test WHERE id = 42","sql.rows_affected":1}`

		actual := buf.Lines()[0]
		if actual != expected {
			t.Fatalf("Expected %s but got %s", expected, actual)
		}
	})
}

func logger() (*gormzap.Logger, *zaptest.Buffer) {
	buf := &zaptest.Buffer{}

	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), buf, zapcore.DebugLevel)
	z := zap.New(core)

	return gormzap.New(z), buf
}
