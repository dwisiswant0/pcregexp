package pcregexp_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/dwisiswant0/pcregexp"
)

func BenchmarkCompile(b *testing.B) {
	patterns := []string{
		`\b\w+@\w+\.\w+\b`,
		`p([a-z]+)ch`,
		`^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$`,
		`(?<=foo)bar`,
		`(\w+)\s+\1`,
	}

	for _, pattern := range patterns {
		b.Run("pcregexp/"+pattern, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				re, _ := pcregexp.Compile(pattern)
				re.Close()
			}
		})

		b.Run("stdlib/"+pattern, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				regexp.Compile(pattern)
			}
		})
	}
}

func BenchmarkMatchString(b *testing.B) {
	tests := []struct {
		name    string
		pattern string
		text    string
	}{
		{"simple", `p([a-z]+)ch`, "peach punch pinch"},
		{"email", `\b\w+@\w+\.\w+\b`, "test@example.com"},
		// {"backreference", `(\w+)\s+\1`, "hello hello world"},
		// {"lookaround", `(?<=foo)bar`, "foobar"},
	}

	for _, tt := range tests {
		pcre := pcregexp.MustCompile(tt.pattern)
		defer pcre.Close()
		re := regexp.MustCompile(tt.pattern)

		b.Run("pcregexp/"+tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pcre.MatchString(tt.text)
			}
		})

		b.Run("stdlib/"+tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				re.MatchString(tt.text)
			}
		})
	}
}

func BenchmarkFind(b *testing.B) {
	tests := []struct {
		name    string
		pattern string
		text    string
	}{
		{"simple", `p([a-z]+)ch`, "peach punch pinch"},
		{"submatch", `(\w+)\s+(\w+)`, "hello world"},
		{"no match", `xyz`, "abc def ghi"},
	}

	for _, tt := range tests {
		pcre := pcregexp.MustCompile(tt.pattern)
		defer pcre.Close()
		re := regexp.MustCompile(tt.pattern)

		b.Run("pcregexp/Find/"+tt.name, func(b *testing.B) {
			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				pcre.Find(data)
			}
		})

		b.Run("stdlib/Find/"+tt.name, func(b *testing.B) {
			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				re.Find(data)
			}
		})

		b.Run("pcregexp/FindString/"+tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pcre.FindString(tt.text)
			}
		})

		b.Run("stdlib/FindString/"+tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				re.FindString(tt.text)
			}
		})
	}
}

func BenchmarkReplace(b *testing.B) {
	tests := []struct {
		name    string
		pattern string
		text    string
		repl    string
	}{
		{"simple", `p([a-z]+)ch`, "peach punch pinch", "FRUIT"},
		{"no match", `xyz`, "abc def ghi", "NONE"},
		{"multiple", `\b\w+\b`, "one two three", "word"},
	}

	for _, tt := range tests {
		pcre := pcregexp.MustCompile(tt.pattern)
		defer pcre.Close()
		re := regexp.MustCompile(tt.pattern)

		b.Run("pcregexp/ReplaceAll/"+tt.name, func(b *testing.B) {
			src := []byte(tt.text)
			repl := []byte(tt.repl)
			for i := 0; i < b.N; i++ {
				pcre.ReplaceAll(src, repl)
			}
		})

		b.Run("stdlib/ReplaceAll/"+tt.name, func(b *testing.B) {
			src := []byte(tt.text)
			repl := []byte(tt.repl)
			for i := 0; i < b.N; i++ {
				re.ReplaceAll(src, repl)
			}
		})

		b.Run("pcregexp/ReplaceAllString/"+tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pcre.ReplaceAllString(tt.text, tt.repl)
			}
		})

		b.Run("stdlib/ReplaceAllString/"+tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				re.ReplaceAllString(tt.text, tt.repl)
			}
		})
	}
}

func BenchmarkFindAll(b *testing.B) {
	tests := []struct {
		name    string
		pattern string
		text    string
	}{
		{"simple", `p([a-z]+)ch`, "peach punch pinch"},
		{"complex", `\b\w+\b`, "one two three four five"},
	}

	for _, tt := range tests {
		pcre := pcregexp.MustCompile(tt.pattern)
		defer pcre.Close()
		re := regexp.MustCompile(tt.pattern)

		b.Run("pcregexp/FindAll/"+tt.name, func(b *testing.B) {
			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				pcre.FindAll(data, -1)
			}
		})

		b.Run("stdlib/FindAll/"+tt.name, func(b *testing.B) {
			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				re.FindAll(data, -1)
			}
		})

		b.Run("pcregexp/FindAllIndex/"+tt.name, func(b *testing.B) {
			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				pcre.FindAllIndex(data, -1)
			}
		})

		b.Run("stdlib/FindAllIndex/"+tt.name, func(b *testing.B) {
			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				re.FindAllIndex(data, -1)
			}
		})
	}
}

func BenchmarkFindAllSubmatch(b *testing.B) {
	tests := []struct {
		name    string
		pattern string
		text    string
	}{
		{"simple", `p([a-z]+)ch`, "peach punch pinch"},
		{"complex", `(\w+)\s+(\w+)`, "hello world goodbye planet"},
	}

	for _, tt := range tests {
		pcre := pcregexp.MustCompile(tt.pattern)
		defer pcre.Close()
		re := regexp.MustCompile(tt.pattern)

		b.Run("pcregexp/FindAllSubmatch/"+tt.name, func(b *testing.B) {
			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				pcre.FindAllSubmatch(data, -1)
			}
		})

		b.Run("stdlib/FindAllSubmatch/"+tt.name, func(b *testing.B) {
			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				re.FindAllSubmatch(data, -1)
			}
		})
	}
}

func BenchmarkExpand(b *testing.B) {
	pattern := `(\w+)\s+(\w+)`
	text := "hello world"
	template := "$2 $1"

	pcre := pcregexp.MustCompile(pattern)
	defer pcre.Close()
	re := regexp.MustCompile(pattern)

	pcreMatch := pcre.FindStringSubmatchIndex(text)
	reMatch := re.FindStringSubmatchIndex(text)

	b.Run("pcregexp/ExpandString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pcre.ExpandString(nil, template, text, pcreMatch)
		}
	})

	b.Run("stdlib/ExpandString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			re.ExpandString(nil, template, text, reMatch)
		}
	})
}

func BenchmarkMarshal(b *testing.B) {
	pattern := `p([a-z]+)ch`
	pcre := pcregexp.MustCompile(pattern)
	defer pcre.Close()
	re := regexp.MustCompile(pattern)

	b.Run("pcregexp/Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pcre.MarshalText()
		}
	})

	b.Run("stdlib/Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			re.MarshalText()
		}
	})
}

func BenchmarkUnmarshal(b *testing.B) {
	pattern := []byte(`p([a-z]+)ch`)

	b.Run("pcregexp/Unmarshal", func(b *testing.B) {
		var re pcregexp.PCREgexp
		for i := 0; i < b.N; i++ {
			re.UnmarshalText(pattern)
		}
	})

	b.Run("stdlib/Unmarshal", func(b *testing.B) {
		var re regexp.Regexp
		for i := 0; i < b.N; i++ {
			re.UnmarshalText(pattern)
		}
	})
}

func BenchmarkRuneReader(b *testing.B) {
	pattern := `p([a-z]+)ch`
	text := "peach punch pinch"

	pcre := pcregexp.MustCompile(pattern)
	defer pcre.Close()
	re := regexp.MustCompile(pattern)

	b.Run("pcregexp/MatchReader", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(text)
			pcre.MatchReader(reader)
		}
	})

	b.Run("stdlib/MatchReader", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(text)
			re.MatchReader(reader)
		}
	})

	b.Run("pcregexp/FindReaderIndex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(text)
			pcre.FindReaderIndex(reader)
		}
	})

	b.Run("stdlib/FindReaderIndex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(text)
			re.FindReaderIndex(reader)
		}
	})
}
