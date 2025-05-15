package pcregexp_test

import (
	"testing"

	"github.com/dwisiswant0/pcregexp"
)

func TestJITCompilation(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		text    string
		want    bool
	}{
		{
			name:    "basic pattern",
			pattern: `p([a-z]+)ch`,
			text:    "peach",
			want:    true,
		},
		{
			name:    "complex pattern",
			pattern: `(?<=foo)bar(?=baz)`,
			text:    "foobarbaz",
			want:    true,
		},
		{
			name:    "backreference",
			pattern: `(\w+)\s+\1`,
			text:    "hello hello world",
			want:    true,
		},
	}

	originalStartSize := pcregexp.DefaultJITStackStartSize
	originalMaxSize := pcregexp.DefaultJITStackMaxSize

	// Test with different JIT stack sizes
	testSizes := []struct {
		name      string
		startSize uint64
		maxSize   uint64
	}{
		{
			name:      "default sizes",
			startSize: pcregexp.DefaultJITStackStartSize,
			maxSize:   pcregexp.DefaultJITStackMaxSize,
		},
		{
			name:      "small stack",
			startSize: 4 * 1024,  // 4KB
			maxSize:   16 * 1024, // 16KB
		},
		{
			name:      "large stack",
			startSize: 64 * 1024,   // 64KB
			maxSize:   1024 * 1024, // 1MB
		},
	}

	for _, size := range testSizes {
		t.Run(size.name, func(t *testing.T) {
			pcregexp.SetJITStackSize(size.startSize, size.maxSize)

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					re := pcregexp.MustCompile(tt.pattern)
					defer re.Close()

					if got := re.MatchString(tt.text); got != tt.want {
						t.Errorf("MatchString() with pattern %q = %v, want %v", tt.pattern, got, tt.want)
					}
				})
			}
		})
	}

	// Restore original stack sizes
	pcregexp.SetJITStackSize(originalStartSize, originalMaxSize)
}

func TestJITPerformance(t *testing.T) {
	pcregexp.SetJITStackSize(64*1024, 1024*1024) // Use optimal stack sizes

	patterns := []struct {
		pattern string
		input   string
	}{
		{
			pattern: `\b\w+@\w+\.\w+\b`,
			input:   "test@example.com",
		},
		{
			pattern: `\d{4}-\d{2}-\d{2}`,
			input:   "2025-05-15",
		},
	}

	for _, p := range patterns {
		t.Run(p.pattern, func(t *testing.T) {
			re := pcregexp.MustCompile(p.pattern)
			defer re.Close()

			// Warm up JIT
			if !re.MatchString(p.input) {
				t.Fatalf("Pattern failed to match during warmup: %s", p.pattern)
			}

			// Test performance
			for i := 0; i < 1000; i++ {
				if !re.MatchString(p.input) {
					t.Errorf("Pattern failed to match on iteration %d: %s", i, p.pattern)
					break
				}
			}
		})
	}
}

func TestJITStackExhaustion(t *testing.T) {
	// Test with very small stack to ensure graceful handling
	pcregexp.SetJITStackSize(1024, 2048)
	pattern := `(?:(^|[^\\])(?:\\\\)*)((?:(")(?:[^"\\]|\\.|\\\\)*")|(?:(')(?:[^'\\]|\\.|\\\\)*')|(?:\/{2}[^\r\n]*)|(?:\/\*[^*]*\*+(?:[^*\/][^*]*\*+)*\/))`
	input := `"""""""""""""""""""""""""""""""` // Many nested quotes

	re := pcregexp.MustCompile(pattern)
	defer re.Close()

	// Should not panic even with small stack
	re.MatchString(input)
}

func TestJITMultiThreading(t *testing.T) {
	pattern := `\b\w+@\w+\.\w+\b`
	input := "test@example.com"
	iterations := 100
	threads := 4

	done := make(chan bool)
	for i := 0; i < threads; i++ {
		go func() {
			re := pcregexp.MustCompile(pattern)
			defer re.Close()

			for j := 0; j < iterations; j++ {
				if !re.MatchString(input) {
					t.Error("Expected match but got none")
				}
			}
			done <- true
		}()
	}

	for i := 0; i < threads; i++ {
		<-done
	}
}
