package gomailparse

import (
	"strings"
	"testing"
)

func TestSimpleEmail(t *testing.T) {
	raw := "From: sender@example.com\r\n" +
		"Subject: Hello\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" +
		"Hello, World!\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if got := part.Header.Get("From"); got != "sender@example.com" {
		t.Errorf("From = %q", got)
	}
	if got := part.Header.Get("Subject"); got != "Hello" {
		t.Errorf("Subject = %q", got)
	}
	if part.ContentType != "text/plain" {
		t.Errorf("ContentType = %q", part.ContentType)
	}
	if part.Charset != "utf-8" {
		t.Errorf("Charset = %q", part.Charset)
	}
	if part.TransferEncoding != "7bit" {
		t.Errorf("TransferEncoding = %q", part.TransferEncoding)
	}

	body := raw[part.BodyPos:part.EndPos]
	if body != "Hello, World!\r\n" {
		t.Errorf("body = %q", body)
	}
}

func TestMultipart(t *testing.T) {
	raw := "Content-Type: multipart/mixed; boundary=\"abc\"\r\n" +
		"\r\n" +
		"--abc\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Hello\r\n" +
		"--abc\r\n" +
		"Content-Type: text/html\r\n" +
		"\r\n" +
		"<p>Hello</p>\r\n" +
		"--abc--\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if part.ContentType != "multipart/mixed" {
		t.Errorf("ContentType = %q", part.ContentType)
	}
	if part.Boundary != "abc" {
		t.Errorf("Boundary = %q", part.Boundary)
	}
	if len(part.Children) != 2 {
		t.Fatalf("len(Children) = %d, want 2", len(part.Children))
	}

	child1 := part.Children[0]
	if child1.ContentType != "text/plain" {
		t.Errorf("child1 ContentType = %q", child1.ContentType)
	}
	if body := raw[child1.BodyPos:child1.EndPos]; body != "Hello" {
		t.Errorf("child1 body = %q, want %q", body, "Hello")
	}

	child2 := part.Children[1]
	if child2.ContentType != "text/html" {
		t.Errorf("child2 ContentType = %q", child2.ContentType)
	}
	if body := raw[child2.BodyPos:child2.EndPos]; body != "<p>Hello</p>" {
		t.Errorf("child2 body = %q, want %q", body, "<p>Hello</p>")
	}
}

func TestNestedMultipart(t *testing.T) {
	raw := "Content-Type: multipart/mixed; boundary=outer\r\n" +
		"\r\n" +
		"--outer\r\n" +
		"Content-Type: multipart/alternative; boundary=inner\r\n" +
		"\r\n" +
		"--inner\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Plain text\r\n" +
		"--inner\r\n" +
		"Content-Type: text/html\r\n" +
		"\r\n" +
		"<p>HTML</p>\r\n" +
		"--inner--\r\n" +
		"--outer\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Attachment\r\n" +
		"--outer--\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if len(part.Children) != 2 {
		t.Fatalf("outer children = %d, want 2", len(part.Children))
	}

	alt := part.Children[0]
	if alt.ContentType != "multipart/alternative" {
		t.Errorf("alt ContentType = %q", alt.ContentType)
	}
	if len(alt.Children) != 2 {
		t.Fatalf("inner children = %d, want 2", len(alt.Children))
	}

	if body := raw[alt.Children[0].BodyPos:alt.Children[0].EndPos]; body != "Plain text" {
		t.Errorf("plain body = %q", body)
	}
	if body := raw[alt.Children[1].BodyPos:alt.Children[1].EndPos]; body != "<p>HTML</p>" {
		t.Errorf("html body = %q", body)
	}

	attach := part.Children[1]
	if body := raw[attach.BodyPos:attach.EndPos]; body != "Attachment" {
		t.Errorf("attach body = %q", body)
	}
}

func TestAttachment(t *testing.T) {
	raw := "Content-Type: multipart/mixed; boundary=bound\r\n" +
		"\r\n" +
		"--bound\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Hello\r\n" +
		"--bound\r\n" +
		"Content-Type: application/pdf; name=\"doc.pdf\"\r\n" +
		"Content-Disposition: attachment; filename=\"document.pdf\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"\r\n" +
		"SGVsbG8=\r\n" +
		"--bound--\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if len(part.Children) != 2 {
		t.Fatalf("children = %d, want 2", len(part.Children))
	}

	att := part.Children[1]
	if att.ContentType != "application/pdf" {
		t.Errorf("ContentType = %q", att.ContentType)
	}
	if att.Name != "doc.pdf" {
		t.Errorf("Name = %q", att.Name)
	}
	if att.Filename != "document.pdf" {
		t.Errorf("Filename = %q", att.Filename)
	}
	if att.ContentDisposition != "attachment" {
		t.Errorf("ContentDisposition = %q", att.ContentDisposition)
	}
	if att.TransferEncoding != "base64" {
		t.Errorf("TransferEncoding = %q", att.TransferEncoding)
	}
	if body := raw[att.BodyPos:att.EndPos]; body != "SGVsbG8=" {
		t.Errorf("body = %q, want %q", body, "SGVsbG8=")
	}
}

func TestHeaderFolding(t *testing.T) {
	raw := "Subject: This is a\r\n" +
		" long subject line\r\n" +
		"From: test@example.com\r\n" +
		"\r\n" +
		"Body\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if got := part.Header.Get("Subject"); got != "This is a long subject line" {
		t.Errorf("Subject = %q", got)
	}
	if got := part.Header.Get("From"); got != "test@example.com" {
		t.Errorf("From = %q", got)
	}
}

func TestDuplicateHeaders(t *testing.T) {
	raw := "Received: from a\r\n" +
		"Received: from b\r\n" +
		"\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	vals := part.Header.Values("Received")
	if len(vals) != 2 {
		t.Fatalf("len(Received) = %d, want 2", len(vals))
	}
	if vals[0] != "from a" || vals[1] != "from b" {
		t.Errorf("Received = %v", vals)
	}
}

func TestEmptyBody(t *testing.T) {
	raw := "Content-Type: text/plain\r\n\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if part.BodySize() != 0 {
		t.Errorf("BodySize = %d, want 0", part.BodySize())
	}
}

func TestLFLineEndings(t *testing.T) {
	raw := "Content-Type: text/plain\n\nHello\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	body := raw[part.BodyPos:part.EndPos]
	if body != "Hello\n" {
		t.Errorf("body = %q", body)
	}
}

func TestLFMultipart(t *testing.T) {
	raw := "Content-Type: multipart/mixed; boundary=b\n" +
		"\n" +
		"--b\n" +
		"Content-Type: text/plain\n" +
		"\n" +
		"Hello\n" +
		"--b--\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if len(part.Children) != 1 {
		t.Fatalf("children = %d, want 1", len(part.Children))
	}
	body := raw[part.Children[0].BodyPos:part.Children[0].EndPos]
	if body != "Hello" {
		t.Errorf("body = %q, want %q", body, "Hello")
	}
}

func TestWalkAndParts(t *testing.T) {
	raw := "Content-Type: multipart/mixed; boundary=b\r\n" +
		"\r\n" +
		"--b\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"A\r\n" +
		"--b\r\n" +
		"Content-Type: text/html\r\n" +
		"\r\n" +
		"B\r\n" +
		"--b--\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	parts := part.Parts()
	if len(parts) != 3 {
		t.Fatalf("len(Parts()) = %d, want 3", len(parts))
	}

	expected := []string{"multipart/mixed", "text/plain", "text/html"}
	for i, want := range expected {
		if parts[i].ContentType != want {
			t.Errorf("parts[%d].ContentType = %q, want %q", i, parts[i].ContentType, want)
		}
	}
}

func TestMultilineBody(t *testing.T) {
	raw := "Content-Type: multipart/mixed; boundary=b\r\n" +
		"\r\n" +
		"--b\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Line 1\r\n" +
		"Line 2\r\n" +
		"Line 3\r\n" +
		"--b--\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	child := part.Children[0]
	body := raw[child.BodyPos:child.EndPos]
	if body != "Line 1\r\nLine 2\r\nLine 3" {
		t.Errorf("body = %q", body)
	}
}

func TestPreambleAndEpilogue(t *testing.T) {
	raw := "Content-Type: multipart/mixed; boundary=b\r\n" +
		"\r\n" +
		"This is the preamble.\r\n" +
		"--b\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Body\r\n" +
		"--b--\r\n" +
		"This is the epilogue.\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if len(part.Children) != 1 {
		t.Fatalf("children = %d, want 1", len(part.Children))
	}
	body := raw[part.Children[0].BodyPos:part.Children[0].EndPos]
	if body != "Body" {
		t.Errorf("body = %q, want %q", body, "Body")
	}
}

func TestBodyReader(t *testing.T) {
	raw := "Content-Type: text/plain\r\n" +
		"\r\n" +
		"Hello, World!\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	ra := strings.NewReader(raw)
	sr := part.BodyReader(ra)
	buf := make([]byte, sr.Size())
	n, err := sr.ReadAt(buf, 0)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf[:n]) != "Hello, World!\r\n" {
		t.Errorf("BodyReader content = %q", buf[:n])
	}
}

func TestCaseInsensitiveHeaderLookup(t *testing.T) {
	raw := "Content-Type: text/plain\r\n\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	// textproto.MIMEHeader.Get canonicalizes keys automatically.
	if part.Header.Get("content-type") != "text/plain" {
		t.Errorf("lowercase lookup failed: %q", part.Header.Get("content-type"))
	}
	if part.Header.Get("CONTENT-TYPE") != "text/plain" {
		t.Errorf("uppercase lookup failed: %q", part.Header.Get("CONTENT-TYPE"))
	}
	// Direct map access with canonical key.
	if vals := part.Header["Content-Type"]; len(vals) != 1 || vals[0] != "text/plain" {
		t.Errorf("direct map access failed: %v", vals)
	}
}

func TestNoContentType(t *testing.T) {
	raw := "From: test@example.com\r\n\r\nHello\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if part.ContentType != "text/plain" {
		t.Errorf("ContentType = %q, want text/plain", part.ContentType)
	}
	if part.Charset != "us-ascii" {
		t.Errorf("Charset = %q, want us-ascii", part.Charset)
	}
}

func TestUnquotedBoundary(t *testing.T) {
	raw := "Content-Type: multipart/mixed; boundary=simple-boundary\r\n" +
		"\r\n" +
		"--simple-boundary\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Content\r\n" +
		"--simple-boundary--\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if part.Boundary != "simple-boundary" {
		t.Errorf("Boundary = %q", part.Boundary)
	}
	if len(part.Children) != 1 {
		t.Fatalf("children = %d, want 1", len(part.Children))
	}
	body := raw[part.Children[0].BodyPos:part.Children[0].EndPos]
	if body != "Content" {
		t.Errorf("body = %q", body)
	}
}

func TestRFC2047EncodedFilename(t *testing.T) {
	raw := "Content-Type: multipart/mixed; boundary=b\r\n" +
		"\r\n" +
		"--b\r\n" +
		"Content-Type: application/pdf;\r\n" +
		" name=\"=?UTF-8?B?5pel5pys6Kqe44OV44Kh44Kk44OrLnBkZg==?=\"\r\n" +
		"Content-Disposition: attachment;\r\n" +
		" filename=\"=?UTF-8?B?5pel5pys6Kqe44OV44Kh44Kk44OrLnBkZg==?=\"\r\n" +
		"\r\n" +
		"data\r\n" +
		"--b--\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	att := part.Children[0]
	want := "\u65e5\u672c\u8a9e\u30d5\u30a1\u30a4\u30eb.pdf" // 日本語ファイル.pdf
	if att.Name != want {
		t.Errorf("Name = %q, want %q", att.Name, want)
	}
	if att.Filename != want {
		t.Errorf("Filename = %q, want %q", att.Filename, want)
	}
}

func TestRFC2047QEncodedFilename(t *testing.T) {
	raw := "Content-Type: multipart/mixed; boundary=b\r\n" +
		"\r\n" +
		"--b\r\n" +
		"Content-Type: text/plain; name=\"=?ISO-8859-1?Q?R=E9sum=E9.txt?=\"\r\n" +
		"Content-Disposition: attachment; filename=\"=?ISO-8859-1?Q?R=E9sum=E9.txt?=\"\r\n" +
		"\r\n" +
		"data\r\n" +
		"--b--\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	att := part.Children[0]
	want := "R\u00e9sum\u00e9.txt" // Résumé.txt
	if att.Name != want {
		t.Errorf("Name = %q, want %q", att.Name, want)
	}
	if att.Filename != want {
		t.Errorf("Filename = %q, want %q", att.Filename, want)
	}
}

func TestDecodedHeader(t *testing.T) {
	raw := "From: =?UTF-8?B?5bGx55Sw5aSq6YOO?= <taro@example.com>\r\n" +
		"Subject: =?UTF-8?Q?Caf=C3=A9?=\r\n" +
		"\r\n"

	part, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	// Raw header is preserved.
	if got := part.Header.Get("Subject"); got != "=?UTF-8?Q?Caf=C3=A9?=" {
		t.Errorf("raw Subject = %q", got)
	}
	// DecodedHeader returns decoded text.
	if got := part.DecodedHeader("Subject"); got != "Caf\u00e9" {
		t.Errorf("decoded Subject = %q, want %q", got, "Caf\u00e9")
	}
	if got := part.DecodedHeader("From"); got != "\u5c71\u7530\u592a\u90ce <taro@example.com>" {
		t.Errorf("decoded From = %q", got)
	}
	// Missing header returns empty.
	if got := part.DecodedHeader("X-Missing"); got != "" {
		t.Errorf("missing header = %q", got)
	}
}
