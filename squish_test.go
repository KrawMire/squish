package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestSquish(t *testing.T) {
	tests := []struct {
		name  string
		input string
		sep   string
		raw   bool
		want  string
	}{
		{
			name:  "json object",
			input: "{\n  \"a\": 1,\n  \"b\": \"x  y\"\n}\n",
			sep:   " ",
			want:  `{"a":1,"b":"x  y"}`,
		},
		{
			name:  "json array",
			input: "[\n  1,\n  2\n]",
			sep:   " ",
			want:  `[1,2]`,
		},
		{
			name:  "raw json",
			input: "{\n  \"a\": 1\n}",
			sep:   " ",
			raw:   true,
			want:  `{ "a": 1 }`,
		},
		{
			name:  "shell continuation",
			input: "docker run \\\n  -it \\\n  ubuntu\n",
			sep:   " ",
			want:  "docker run -it ubuntu",
		},
		{
			name:  "quotes keep internal spaces",
			input: "echo \"hello   world\" \\\n  foo\n",
			sep:   " ",
			want:  `echo "hello   world" foo`,
		},
		{
			name:  "single quotes ignore backslash",
			input: "echo 'a  b\\'\n c",
			sep:   " ",
			want:  `echo 'a  b\' c`,
		},
		{
			name:  "newline inside quotes",
			input: "echo \"hello\nworld\"",
			sep:   " ",
			want:  `echo "hello world"`,
		},
		{
			name:  "continuation without space concatenates",
			input: "foo\\\nbar",
			sep:   " ",
			want:  "foobar",
		},
		{
			name:  "escaped backslash then continuation",
			input: "foo " + `\\\` + "\nbar",
			sep:   " ",
			want:  `foo \bar`,
		},
		{
			name:  "even backslashes are literal",
			input: "foo " + `\\` + "\nbar",
			sep:   " ",
			want:  `foo \\ bar`,
		},
		{
			name:  "literal trailing backslashes",
			input: `foo \\`,
			sep:   " ",
			want:  `foo \\`,
		},
		{
			name:  "sql",
			input: "SELECT *\n  FROM users\n WHERE id = 1",
			sep:   " ",
			want:  "SELECT * FROM users WHERE id = 1",
		},
		{
			name:  "unicode",
			input: "héllo\nworld",
			sep:   " ",
			want:  "héllo world",
		},
		{
			name:  "blank",
			input: "  \n\t\n ",
			sep:   " ",
			want:  "",
		},
		{
			name:  "custom separator",
			input: "a \\\n b \n\n c",
			sep:   "; ",
			want:  "a; b; c",
		},
		{
			name:  "json with trailing junk stays text",
			input: "{\"a\": 1}\nmore",
			sep:   " ",
			want:  `{"a": 1} more`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := squish(tt.input, tt.sep, tt.raw)
			if got != tt.want {
				t.Fatalf("squish() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"hello\nworld"}, strings.NewReader(""), &out, &errOut)
	if code != 0 {
		t.Fatalf("exit %d, stderr %s", code, errOut.String())
	}
	if out.String() != "hello world\n" {
		t.Fatalf("stdout = %q", out.String())
	}

	out.Reset()
	errOut.Reset()
	code = run(nil, strings.NewReader("{\n  \"ok\": true\n}\n"), &out, &errOut)
	if code != 0 {
		t.Fatalf("exit %d, stderr %s", code, errOut.String())
	}
	if out.String() != "{\"ok\":true}\n" {
		t.Fatalf("stdout = %q", out.String())
	}

	out.Reset()
	errOut.Reset()
	code = run([]string{"-h"}, strings.NewReader(""), &out, &errOut)
	if code != 0 {
		t.Fatalf("help exit %d", code)
	}
	if !strings.Contains(errOut.String(), "squish turns") {
		t.Fatalf("help = %q", errOut.String())
	}
}
