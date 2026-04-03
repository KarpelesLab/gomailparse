package gomime

import (
	"io"
	"strings"
)

// Part represents a MIME part with headers and byte offset information.
// Body content is not stored; use the offsets with the original data
// source to extract body content afterward.
//
// Offsets use half-open intervals: body content spans [BodyPos, EndPos).
type Part struct {
	// Byte offsets into the original stream.
	StartPos int64 // start of this part (beginning of headers)
	BodyPos  int64 // start of body (after headers and blank line separator)
	EndPos   int64 // end of body content

	// Content metadata extracted from headers.
	ContentType        string // lowercase, e.g. "text/plain", "multipart/mixed"
	Boundary           string // boundary parameter for multipart types
	Charset            string // charset parameter, defaults to "us-ascii"
	TransferEncoding   string // Content-Transfer-Encoding, defaults to "7bit"
	ContentDisposition string // "attachment", "inline", etc.
	Name               string // name parameter from Content-Type
	Filename           string // filename parameter from Content-Disposition

	// Children contains sub-parts for multipart types.
	Children []*Part

	headers []headerEntry
}

// HeaderField is a single header key-value pair as returned by [Part.Headers].
type HeaderField struct {
	Key   string // original case
	Value string // unfolded value
}

type headerEntry struct {
	key   string // original case
	value string // unfolded value
}

// Get returns the first header value matching key (case-insensitive).
// Returns empty string if not found.
func (p *Part) Get(key string) string {
	for _, h := range p.headers {
		if strings.EqualFold(h.key, key) {
			return h.value
		}
	}
	return ""
}

// GetAll returns all header values matching key (case-insensitive).
func (p *Part) GetAll(key string) []string {
	var result []string
	for _, h := range p.headers {
		if strings.EqualFold(h.key, key) {
			result = append(result, h.value)
		}
	}
	return result
}

// Headers returns all headers in their original order.
func (p *Part) Headers() []HeaderField {
	result := make([]HeaderField, len(p.headers))
	for i, h := range p.headers {
		result[i] = HeaderField{Key: h.key, Value: h.value}
	}
	return result
}

// Walk calls fn for each part in the MIME tree in depth-first order.
// If fn returns false, traversal stops and Walk returns false.
func (p *Part) Walk(fn func(*Part) bool) bool {
	if !fn(p) {
		return false
	}
	for _, child := range p.Children {
		if !child.Walk(fn) {
			return false
		}
	}
	return true
}

// Parts returns a flat list of all parts in depth-first order.
func (p *Part) Parts() []*Part {
	var result []*Part
	p.Walk(func(part *Part) bool {
		result = append(result, part)
		return true
	})
	return result
}

// BodySize returns the size of the body in bytes.
func (p *Part) BodySize() int64 {
	return p.EndPos - p.BodyPos
}

// BodyReader returns an [io.SectionReader] for this part's body content.
// The provided [io.ReaderAt] should be the original data source that
// was parsed (e.g. an [*os.File] or [*bytes.Reader]).
func (p *Part) BodyReader(ra io.ReaderAt) *io.SectionReader {
	return io.NewSectionReader(ra, p.BodyPos, p.EndPos-p.BodyPos)
}
