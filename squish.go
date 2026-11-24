package main

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// squish turns multi-line text into a single line.
//
// A JSON object or array is compacted, unless raw is set. Any other text has
// shell line-continuations removed and whitespace outside quotes folded into
// sep. Inside quotes, repeated spaces stay, and line breaks become one space.
// When sep is not a single space, trimmed lines are joined with sep instead.
func squish(input, sep string, raw bool) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.ReplaceAll(input, "\r", "\n")
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	if !raw && sep == " " {
		if compact, ok := compactJSON(trimmed); ok {
			return compact
		}
		return collapse(trimmed)
	}
	return joinLines(trimmed, sep)
}

func compactJSON(s string) (string, bool) {
	if s == "" || (s[0] != '{' && s[0] != '[') {
		return "", false
	}
	dec := json.NewDecoder(strings.NewReader(s))
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return "", false
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return "", false
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return "", false
	}
	return buf.String(), true
}

func joinLines(s, sep string) string {
	lines := strings.Split(s, "\n")
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		line = stripContinuation(line)
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts = append(parts, line)
	}
	return strings.Join(parts, sep)
}

// stripContinuation removes one trailing backslash when the line ends with an
// odd run of them, which is a shell line continuation.
func stripContinuation(line string) string {
	n := 0
	for n < len(line) && line[len(line)-1-n] == '\\' {
		n++
	}
	if n%2 == 1 {
		line = line[:len(line)-1]
	}
	return line
}

func collapse(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	pendingSpace := false
	quote := byte(0)
	last := byte(0)

	write := func(c byte) {
		b.WriteByte(c)
		last = c
		pendingSpace = false
	}
	space := func() {
		if b.Len() > 0 && last != ' ' {
			pendingSpace = true
		}
	}
	flushSpace := func() {
		if pendingSpace {
			write(' ')
		}
	}

	for i := 0; i < len(s); {
		c := s[i]
		if quote != 0 {
			if c == '\n' {
				space()
				if pendingSpace && (b.Len() == 0 || last != ' ') {
					write(' ')
				}
				pendingSpace = false
				i++
				continue
			}
			if quote == '"' && c == '\\' && i+1 < len(s) {
				flushSpace()
				write(c)
				write(s[i+1])
				i += 2
				continue
			}
			if c == quote {
				flushSpace()
				write(c)
				quote = 0
				i++
				continue
			}
			flushSpace()
			write(c)
			i++
			continue
		}

		if c == '"' || c == '\'' || c == '`' {
			flushSpace()
			quote = c
			write(c)
			i++
			continue
		}

		if c == '\\' {
			n, next := backslashRun(s, i)
			for range n {
				flushSpace()
				write('\\')
			}
			i = next
			continue
		}

		if c < utf8.RuneSelf {
			if isASCIISpace(c) {
				space()
				i++
				continue
			}
			flushSpace()
			write(c)
			i++
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		if unicode.IsSpace(r) {
			space()
			i += size
			continue
		}
		flushSpace()
		b.WriteString(s[i : i+size])
		last = s[i+size-1]
		pendingSpace = false
		i += size
	}
	return b.String()
}

func isASCIISpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\f'
}

// backslashRun consumes the backslashes at i. An odd run that escapes a
// following newline is a shell continuation: the escaped backslash and the
// newline are dropped, and n is the number of literal backslashes to keep.
// Otherwise every backslash is literal and next points at the following byte.
func backslashRun(s string, i int) (n, next int) {
	j := i
	for j < len(s) && s[j] == '\\' {
		j++
	}
	n = j - i
	if n%2 == 1 && j < len(s) && s[j] == '\n' {
		return n / 2, j + 1
	}
	return n, j
}
