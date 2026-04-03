# gomailparse

[![Go Reference](https://pkg.go.dev/badge/github.com/KarpelesLab/gomailparse.svg)](https://pkg.go.dev/github.com/KarpelesLab/gomailparse)
[![Build Status](https://github.com/KarpelesLab/gomailparse/actions/workflows/test.yml/badge.svg)](https://github.com/KarpelesLab/gomailparse/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/KarpelesLab/gomailparse/badge.svg?branch=master)](https://coveralls.io/github/KarpelesLab/gomailparse?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/KarpelesLab/gomailparse)](https://goreportcard.com/report/github.com/KarpelesLab/gomailparse)

A streaming MIME parser for Go. Reads emails from an `io.Reader` and produces a part tree containing only headers and byte offsets — body data is never buffered. This makes it efficient for indexing large messages and locating attachments or part bodies for later random access.

## Features

- **Stream-based** — parses from any `io.Reader` in a single pass
- **Offset-only** — stores byte positions (`StartPos`, `BodyPos`, `EndPos`) instead of body content
- **No regexp** — all parsing is done with straightforward byte/string operations
- **Nested multipart** — handles arbitrary nesting of multipart types
- **RFC compliant** — handles header folding (RFC 2822), boundary delimiters (RFC 2046), both CRLF and LF line endings

## Install

```
go get github.com/KarpelesLab/gomailparse
```

## Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/KarpelesLab/gomailparse"
)

func main() {
	f, _ := os.Open("message.eml")
	defer f.Close()

	part, err := gomailparse.Parse(f)
	if err != nil {
		panic(err)
	}

	// Walk all parts
	for _, p := range part.Parts() {
		fmt.Printf("%-30s  body=[%d:%d] (%d bytes)\n",
			p.ContentType, p.BodyPos, p.EndPos, p.BodySize())

		// Read an attachment body using offsets
		if p.ContentDisposition == "attachment" {
			reader := p.BodyReader(f)
			// reader is an *io.SectionReader positioned at the body
			_ = reader
		}
	}
}
```

## API

### `gomailparse.Parse(r io.Reader) (*Part, error)`

Parses a MIME message and returns the root part. Only headers and byte offsets are retained.

### `Part`

```go
type Part struct {
    StartPos int64  // byte offset: start of part (headers)
    BodyPos  int64  // byte offset: start of body
    EndPos   int64  // byte offset: end of body

    Header textproto.MIMEHeader // parsed headers (canonical key form)

    ContentType        string // e.g. "text/plain"
    Boundary           string // for multipart types
    Charset            string // defaults to "us-ascii"
    TransferEncoding   string // defaults to "7bit"
    ContentDisposition string // "attachment", "inline", etc.
    Name               string // name= from Content-Type
    Filename           string // filename= from Content-Disposition (RFC 2231 supported)

    Children []*Part
}
```

### Header access

`Header` is a standard [`textproto.MIMEHeader`](https://pkg.go.dev/net/textproto#MIMEHeader):

- `Header.Get("Subject")` — first value (case-insensitive)
- `Header.Values("Received")` — all values
- `Header["Content-Type"]` — direct map access with canonical key

### Traversal

- `Walk(func(*Part) bool) bool` — depth-first traversal (return false to stop)
- `Parts() []*Part` — flat list of all parts

### Body access

- `BodySize() int64` — body size in bytes
- `BodyReader(ra io.ReaderAt) *io.SectionReader` — positioned reader for the body

## License

MIT — see [LICENSE](LICENSE).
