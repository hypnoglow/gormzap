package gormzap_test

import (
	"testing"
	"time"
)

func BenchmarkLogger_Print(b *testing.B) {
	l, _ := logger()

	for i := 0; i < b.N; i++ {
		l.Print(
			"sql",
			"/some/file.go:34",
			time.Millisecond*5,
			"SELECT * FROM test WHERE id = $1",
			[]interface{}{42},
			int64(1),
		)
	}
}
