# redis-lock

[![Build Status](https://travis-ci.org/bsm/redis-lock.png?branch=master)](https://travis-ci.org/bsm/redis-lock)
[![GoDoc](https://godoc.org/github.com/bsm/redis-lock?status.png)](http://godoc.org/github.com/bsm/redis-lock)
[![Go Report Card](https://goreportcard.com/badge/github.com/bsm/redis-lock)](https://goreportcard.com/report/github.com/bsm/redis-lock)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Simplified distributed locking implementation using [Redis](http://redis.io/topics/distlock).
For more information, please see examples.

## Examples

```go
import (
  "fmt"
  "time"

  "github.com/bsm/redis-lock"
  "github.com/go-redis/redis"
)

func main() {{ "Example" | code }}
```

## Documentation

Full documentation is available on [GoDoc](http://godoc.org/github.com/bsm/redis-lock)

## Testing

Simply run:

    make


