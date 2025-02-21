package regexp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCompile(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
		isPCRE  bool
	}{
		{
			name:    "simple pattern",
			pattern: "hello",
			wantErr: false,
			isPCRE:  false,
		},
		{
			name:    "invalid pattern",
			pattern: "[",
			wantErr: true,
			isPCRE:  false,
		},
		{
			name:    "pcre lookahead",
			pattern: "foo(?=bar)",
			wantErr: false,
			isPCRE:  true,
		},
		{
			name:    "pcre lookbehind",
			pattern: "(?<=foo)bar",
			wantErr: false,
			isPCRE:  true,
		},
		{
			name:    "pcre backreference",
			pattern: "(foo)\\1",
			wantErr: false,
			isPCRE:  true,
		},
		{
			name:    "unicode grapheme",
			pattern: "a\\Xb",
			wantErr: false,
			isPCRE:  true,
		},
		{
			name:    "horizontal whitespace",
			pattern: "a\\hb",
			wantErr: false,
			isPCRE:  true,
		},
		{
			name:    "atomic group",
			pattern: "(?>abc)",
			wantErr: false,
			isPCRE:  true,
		},
		{
			name:    "recursion",
			pattern: "(?R)",
			wantErr: false,
			isPCRE:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := Compile(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if re.IsPCRE() != tt.isPCRE {
				t.Errorf("Compile() isPCRE = %v, want %v", re.IsPCRE(), tt.isPCRE)
			}
		})
	}
}

func TestCompile_CommonWebAttacks(t *testing.T) {
	url := "https://github.com/teler-sh/teler-resources/raw/refs/heads/master/db/common-web-attacks.json"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to fetch JSON from %q: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected HTTP status: %s", resp.Status)
	}

	var data struct {
		Filters []struct {
			ID          int      `json:"id"`
			Rule        string   `json:"rule"`
			Description string   `json:"description"`
			Tags        []string `json:"tags"`
			Impact      int      `json:"impact"`
		} `json:"filters"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	for _, filter := range data.Filters {
		t.Run(fmt.Sprintf("ID-%d", filter.ID), func(t *testing.T) {
			_, err := Compile(filter.Rule)
			if err != nil {
				t.Errorf("Failed to compile rule %q: %v", filter.Rule, err)
				return
			}
		})
	}
}

func TestRegexp_Match(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    bool
	}{
		{
			name:    "simple match",
			pattern: "hello",
			input:   "hello world",
			want:    true,
		},
		{
			name:    "pcre lookahead match",
			pattern: "foo(?=bar)",
			input:   "foobar",
			want:    true,
		},
		{
			name:    "pcre lookahead non-match",
			pattern: "foo(?=bar)",
			input:   "foobaz",
			want:    false,
		},
		{
			name:    "pcre lookbehind match",
			pattern: "(?<=foo)bar",
			input:   "foobar",
			want:    true,
		},
		{
			name:    "pcre backreference match",
			pattern: "(foo)\\1",
			input:   "foofoo",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			defer re.Close()
			if got := re.MatchString(tt.input); got != tt.want {
				t.Errorf("Regexp.MatchString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexp_Find(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    string
	}{
		{
			name:    "simple find",
			pattern: "hello",
			input:   "hello world",
			want:    "hello",
		},
		{
			name:    "pcre lookahead find",
			pattern: "foo(?=bar)",
			input:   "foobar",
			want:    "foo",
		},
		{
			name:    "pcre lookbehind find",
			pattern: "(?<=foo)bar",
			input:   "foobar",
			want:    "bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			defer re.Close()
			if got := re.FindString(tt.input); got != tt.want {
				t.Errorf("Regexp.FindString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexp_Replace(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		replace string
		want    string
	}{
		{
			name:    "simple replace",
			pattern: "hello",
			input:   "hello world",
			replace: "hi",
			want:    "hi world",
		},
		{
			name:    "pcre replace with backreference",
			pattern: "(foo)(bar)",
			input:   "foobar",
			replace: "$2$1",
			want:    "barfoo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			defer re.Close()
			if got := re.ReplaceAllString(tt.input, tt.replace); got != tt.want {
				t.Errorf("Regexp.ReplaceAllString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexp_FindSubmatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    []string
	}{
		{
			name:    "simple submatch",
			pattern: "(hello) (world)",
			input:   "hello world",
			want:    []string{"hello world", "hello", "world"},
		},
		{
			name:    "pcre submatch with lookahead",
			pattern: "(foo)(?=bar)",
			input:   "foobar",
			want:    []string{"foo", "foo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := MustCompile(tt.pattern)
			defer re.Close()
			if got := re.FindStringSubmatch(tt.input); !stringsEqual(got, tt.want) {
				t.Errorf("Regexp.FindStringSubmatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
