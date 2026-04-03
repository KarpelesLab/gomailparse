# gomime

A streaming MIME parser for Go. Reads emails from an `io.Reader` and produces a part tree containing only headers and byte offsets — body data is never buffered. This makes it efficient for indexing large messages and locating attachments or part bodies for later random access.

## Features

- **Stream-based** — parses from any `io.Reader` in a single pass
- **Offset-only** — stores byte positions (`StartPos`, `BodyPos`, `EndPos`) instead of body content
- **No regexp** — all parsing is done with straightforward byte/string operations
- **Nested multipart** — handles arbitrary nesting of multipart types
- **RFC compliant** — handles header folding (RFC 2822), boundary delimiters (RFC 2046), both CRLF and LF line endings

## Install

```
go get github.com/KarpelesLab/gomime
```

## Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/KarpelesLab/gomime"
)

func main() {
	f, _ := os.Open("message.eml")
	defer f.Close()

	part, err := gomime.Parse(f)
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

### `gomime.Parse(r io.Reader) (*Part, error)`

Parses a MIME message and returns the root part. Only headers and byte offsets are retained.

### `Part`

```go
type Part struct {
    StartPos int64  // byte offset: start of part (headers)
    BodyPos  int64  // byte offset: start of body
    EndPos   int64  // byte offset: end of body

    ContentType        string // e.g. "text/plain"
    Boundary           string // for multipart types
    Charset            string // defaults to "us-ascii"
    TransferEncoding   string // defaults to "7bit"
    ContentDisposition string // "attachment", "inline", etc.
    Name               string // name= from Content-Type
    Filename           string // filename= from Content-Disposition

    Children []*Part
}
```

### Header access

- `Get(key string) string` — first value, case-insensitive lookup
- `GetAll(key string) []string` — all values
- `Headers() []HeaderField` — all headers in order

### Traversal

- `Walk(func(*Part) bool) bool` — depth-first traversal (return false to stop)
- `Parts() []*Part` — flat list of all parts

### Body access

- `BodySize() int64` — body size in bytes
- `BodyReader(ra io.ReaderAt) *io.SectionReader` — positioned reader for the body

## License

MIT — see [LICENSE](LICENSE).
