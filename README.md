# gormzap

[![GoDoc](https://godoc.org/github.com/hypnoglow/gormzap?status.svg)](https://godoc.org/github.com/hypnoglow/gormzap)
[![CircleCI](https://circleci.com/gh/hypnoglow/gormzap.svg?style=shield)](https://circleci.com/gh/hypnoglow/gormzap)
[![GitHub release](https://img.shields.io/github/tag/hypnoglow/gormzap.svg)](https://github.com/hypnoglow/gormzap/releases)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

[GORM](https://github.com/jinzhu/gorm) logger implementation using [zap](https://github.com/uber-go/zap).

## Usage

```go
var debug bool // shows if we have debug enabled in our app

db, err := gorm.Open("postgres", dsn)
if err != nil {
    panic(err)
}

if debug {
    // By default, gorm logs only errors. If we set LogMode to true,
    // then all queries will be logged.
    // WARNING: if you explicitly set this to false, then even
    // errors won't be logged.
    db.LogMode(true)
}

log := zap.NewExample()

db.SetLogger(gormzap.New(log, gormzap.WithLevel(zap.DebugLevel)))
```
