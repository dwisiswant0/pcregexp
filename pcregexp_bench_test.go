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
		`(?<=foo)bar`,
	}

	for _, pattern := range patterns {
		b.ResetTimer()
		b.Run("pcregexp/"+pattern, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				re, _ := pcregexp.Compile(pattern)
				re.Close()
			}
		})

		b.ResetTimer()
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
		{"backreference", `(\w+)\s+\1`, "hello hello world"},
		{"lookaround", `(?<=foo)bar`, "foobar"},
	}

	for _, tt := range tests {
		pcre := pcregexp.MustCompile(tt.pattern)
		defer pcre.Close()
		re, _ := regexp.Compile(tt.pattern)

		b.ResetTimer()
		b.Run("pcregexp/"+tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pcre.MatchString(tt.text)
			}
		})

		b.ResetTimer()
		b.Run("stdlib/"+tt.name, func(b *testing.B) {
			if pcregexp.NeedsPCRE(tt.pattern) {
				b.Skip("skipping test for pattern requiring PCRE")
			}

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
		{"backreference", `(\w+)\s+\1`, "hello hello world"},
		{"lookaround", `(?<=foo)bar`, "foobar"},
	}

	for _, tt := range tests {
		pcre := pcregexp.MustCompile(tt.pattern)
		defer pcre.Close()
		re, _ := regexp.Compile(tt.pattern)

		b.ResetTimer()
		b.Run("pcregexp/Find/"+tt.name, func(b *testing.B) {
			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				pcre.Find(data)
			}
		})

		b.ResetTimer()
		b.Run("stdlib/Find/"+tt.name, func(b *testing.B) {
			if pcregexp.NeedsPCRE(tt.pattern) {
				b.Skip("skipping test for pattern requiring PCRE")
			}

			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				re.Find(data)
			}
		})

		b.ResetTimer()
		b.Run("pcregexp/FindString/"+tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pcre.FindString(tt.text)
			}
		})

		b.ResetTimer()
		b.Run("stdlib/FindString/"+tt.name, func(b *testing.B) {
			if pcregexp.NeedsPCRE(tt.pattern) {
				b.Skip("skipping test for pattern requiring PCRE")
			}

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

		b.ResetTimer()
		b.Run("pcregexp/ReplaceAll/"+tt.name, func(b *testing.B) {
			src := []byte(tt.text)
			repl := []byte(tt.repl)
			for i := 0; i < b.N; i++ {
				pcre.ReplaceAll(src, repl)
			}
		})

		b.ResetTimer()
		b.Run("stdlib/ReplaceAll/"+tt.name, func(b *testing.B) {
			src := []byte(tt.text)
			repl := []byte(tt.repl)
			for i := 0; i < b.N; i++ {
				re.ReplaceAll(src, repl)
			}
		})

		b.ResetTimer()
		b.Run("pcregexp/ReplaceAllString/"+tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pcre.ReplaceAllString(tt.text, tt.repl)
			}
		})

		b.ResetTimer()
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
		{"backreference", `(\w+)\s+\1`, "hello hello world"},
		{"lookaround", `(?<=foo)bar`, "foobar"},
	}

	for _, tt := range tests {
		pcre := pcregexp.MustCompile(tt.pattern)
		defer pcre.Close()
		re, _ := regexp.Compile(tt.pattern)

		b.ResetTimer()
		b.Run("pcregexp/FindAll/"+tt.name, func(b *testing.B) {
			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				pcre.FindAll(data, -1)
			}
		})

		b.ResetTimer()
		b.Run("stdlib/FindAll/"+tt.name, func(b *testing.B) {
			if pcregexp.NeedsPCRE(tt.pattern) {
				b.Skip("skipping test for pattern requiring PCRE")
			}

			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				re.FindAll(data, -1)
			}
		})

		b.ResetTimer()
		b.Run("pcregexp/FindAllIndex/"+tt.name, func(b *testing.B) {
			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				pcre.FindAllIndex(data, -1)
			}
		})

		b.ResetTimer()
		b.Run("stdlib/FindAllIndex/"+tt.name, func(b *testing.B) {
			if pcregexp.NeedsPCRE(tt.pattern) {
				b.Skip("skipping test for pattern requiring PCRE")
			}

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
		{"backreference", `(\w+)\s+\1`, "hello hello world"},
		{"lookaround", `(?<=foo)bar`, "foobar"},
	}

	for _, tt := range tests {
		pcre := pcregexp.MustCompile(tt.pattern)
		defer pcre.Close()
		re, _ := regexp.Compile(tt.pattern)

		b.ResetTimer()
		b.Run("pcregexp/FindAllSubmatch/"+tt.name, func(b *testing.B) {
			data := []byte(tt.text)
			for i := 0; i < b.N; i++ {
				pcre.FindAllSubmatch(data, -1)
			}
		})

		b.ResetTimer()
		b.Run("stdlib/FindAllSubmatch/"+tt.name, func(b *testing.B) {
			if pcregexp.NeedsPCRE(tt.pattern) {
				b.Skip("skipping test for pattern requiring PCRE")
			}

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

	b.ResetTimer()
	b.Run("pcregexp/ExpandString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pcre.ExpandString(nil, template, text, pcreMatch)
		}
	})

	b.ResetTimer()
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

	b.ResetTimer()
	b.Run("pcregexp/Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pcre.MarshalText()
		}
	})

	b.ResetTimer()
	b.Run("stdlib/Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			re.MarshalText()
		}
	})
}

func BenchmarkUnmarshal(b *testing.B) {
	pattern := []byte(`p([a-z]+)ch`)

	b.ResetTimer()
	b.Run("pcregexp/Unmarshal", func(b *testing.B) {
		var re pcregexp.PCREgexp
		for i := 0; i < b.N; i++ {
			re.UnmarshalText(pattern)
		}
	})

	b.ResetTimer()
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

	b.ResetTimer()
	b.Run("pcregexp/MatchReader", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(text)
			pcre.MatchReader(reader)
		}
	})

	b.ResetTimer()
	b.Run("stdlib/MatchReader", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(text)
			re.MatchReader(reader)
		}
	})

	b.ResetTimer()
	b.Run("pcregexp/FindReaderIndex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(text)
			pcre.FindReaderIndex(reader)
		}
	})

	b.ResetTimer()
	b.Run("stdlib/FindReaderIndex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(text)
			re.FindReaderIndex(reader)
		}
	})
}
