// Package regexp provides a drop-in replacement for Go's standard regexp package
// with added support for Perl-compatible regular expressions (PCRE).
//
// This package automatically selects between using the standard library's
// regexp engine and a PCRE-based engine (pcregexp package) based on the
// features used in the regular expression. If the pattern contains PCRE-
// specific constructs such as lookahead/lookbehind assertions or
// backreferences, the PCRE engine is employed; otherwise, the standard library
// implementation is used.
//
// The [Regexp] type represents a compiled regular expression and wraps either a
// standard [regexp.Regexp] or a [pcregexp.PCREgexp], exposing a unified API for
// matching, searching, replacing, and more.
//
// Use this package when you require advanced regex features not supported by
// the standard library's regexp package.
package regexp

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/dwisiswant0/pcregexp"
)

// Regexp is the representation of a compiled regular expression.
type Regexp struct {
	regexp   *regexp.Regexp
	pcregexp *pcregexp.PCREgexp
	pattern  string
}

// IsPCRE reports whether the [Regexp] is a PCRE.
func (r *Regexp) IsPCRE() bool {
	return r.pcregexp != nil
}

// needsPCRE checks if the pattern contains features that require PCRE.
func needsPCRE(pattern string) bool {
	lookarounds := []string{
		"(?=", "(?!", // Positive and negative lookahead.
		"(?<=", "(?<!", // Positive and negative lookbehind.
	}

	unsupportedTokens := []string{
		`\X`,         // Unicode grapheme
		`\h`,         // horizontal whitespace
		`\H`,         // not horizontal whitespace
		`\V`,         // not vertical whitespace
		`\R`,         // line break
		`\N`,         // not line break
		`\Z`,         // end of string
		`\G`,         // previous match end
		`\K`,         // keep out
		`\c`,         // control character escape (e.g. \cI)
		`\e`,         // escape character
		"(?>",        // atomic group
		"(?|",        // branch reset group
		"(?(DEFINE)", // define group
		"(?(",        // conditional group (matches conditionals, but not non-capturing which is "(?:")
		"(?#",        // comment
		"(?R)",       // recursion
	}

	// Check for lookahead and lookbehind assertions.
	for _, token := range lookarounds {
		if strings.Contains(pattern, token) {
			return true
		}
	}

	// Check for unsupported tokens.
	for _, token := range unsupportedTokens {
		if strings.Contains(pattern, token) {
			return true
		}
	}

	// Check for backreferences using simple string matching.
	groups := 0
	escaped := false
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '\\' {
			escaped = !escaped
			continue
		}
		if !escaped && pattern[i] == '(' {
			// Skip non-capturing groups and named groups.
			if i+2 < len(pattern) && pattern[i+1] == '?' {
				if pattern[i+2] == ':' || pattern[i+2] == 'P' {
					continue
				}
			}
			groups++
		}
		escaped = false
	}

	// If we have any capturing groups, look for backreferences.
	if groups > 0 {
		escaped = false
		for i := 0; i < len(pattern); i++ {
			if pattern[i] == '\\' {
				if !escaped && i+1 < len(pattern) {
					next := pattern[i+1]
					if next >= '1' && next <= '9' {
						return true
					}
				}
				escaped = !escaped
			} else {
				escaped = false
			}
		}
	}

	return false
}

func Compile(pattern string) (*Regexp, error) {
	if needsPCRE(pattern) {
		pcre, err := pcregexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		return &Regexp{pattern: pattern, pcregexp: pcre}, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &Regexp{pattern: pattern, regexp: re}, nil
}

func MustCompile(pattern string) *Regexp {
	re, err := Compile(pattern)
	if err != nil {
		v := fmt.Sprintf("regexp: Compile(%q): %s", pattern, err.Error())
		panic(v)
	}
	return re
}

// Match reports whether the byte slice b
// contains any match of the regular expression pattern.
// More complicated queries need to use [Compile] and the full [Regexp] interface.
func Match(pattern string, b []byte) (matched bool, err error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.Match(b), nil
}

// MatchReader reports whether the text returned by the [io.RuneReader]
// contains any match of the regular expression pattern.
// More complicated queries need to use [Compile] and the full [Regexp] interface.
func MatchReader(pattern string, r io.RuneReader) (matched bool, err error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchReader(r), nil
}

// MatchString reports whether the string s
// contains any match of the regular expression pattern.
// More complicated queries need to use [Compile] and the full [Regexp] interface.
func MatchString(pattern string, s string) (matched bool, err error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(s), nil
}

// Close releases any resources used by the [pcregexp.PCREgexp].
func (r *Regexp) Close() {
	if r.pcregexp != nil {
		r.pcregexp.Close()
	}
}

func (r *Regexp) Find(b []byte) []byte {
	if r.pcregexp != nil {
		return r.pcregexp.Find(b)
	}
	return r.regexp.Find(b)
}

func (r *Regexp) FindAll(b []byte, n int) [][]byte {
	if r.pcregexp != nil {
		return r.pcregexp.FindAll(b, n)
	}
	return r.regexp.FindAll(b, n)
}

func (r *Regexp) FindAllIndex(b []byte, n int) [][]int {
	if r.pcregexp != nil {
		return r.pcregexp.FindAllIndex(b, n)
	}
	return r.regexp.FindAllIndex(b, n)
}

func (r *Regexp) FindIndex(b []byte) []int {
	if r.pcregexp != nil {
		return r.pcregexp.FindIndex(b)
	}
	return r.regexp.FindIndex(b)
}

func (r *Regexp) FindString(s string) string {
	if r.pcregexp != nil {
		return r.pcregexp.FindString(s)
	}
	return r.regexp.FindString(s)
}

func (r *Regexp) FindAllString(s string, n int) []string {
	if r.pcregexp != nil {
		return r.pcregexp.FindAllString(s, n)
	}
	return r.regexp.FindAllString(s, n)
}

func (r *Regexp) FindStringIndex(s string) []int {
	if r.pcregexp != nil {
		return r.pcregexp.FindStringIndex(s)
	}
	return r.regexp.FindStringIndex(s)
}

func (r *Regexp) FindAllStringIndex(s string, n int) [][]int {
	if r.pcregexp != nil {
		return r.pcregexp.FindAllStringIndex(s, n)
	}
	return r.regexp.FindAllStringIndex(s, n)
}

func (r *Regexp) FindSubmatch(b []byte) [][]byte {
	if r.pcregexp != nil {
		return r.pcregexp.FindSubmatch(b)
	}
	return r.regexp.FindSubmatch(b)
}

func (r *Regexp) FindStringSubmatch(s string) []string {
	if r.pcregexp != nil {
		return r.pcregexp.FindStringSubmatch(s)
	}
	return r.regexp.FindStringSubmatch(s)
}

func (r *Regexp) FindAllSubmatch(b []byte, n int) [][][]byte {
	if r.pcregexp != nil {
		return r.pcregexp.FindAllSubmatch(b, n)
	}
	return r.regexp.FindAllSubmatch(b, n)
}

func (r *Regexp) FindAllStringSubmatch(s string, n int) [][]string {
	if r.pcregexp != nil {
		return r.pcregexp.FindAllStringSubmatch(s, n)
	}
	return r.regexp.FindAllStringSubmatch(s, n)
}

func (r *Regexp) Match(b []byte) bool {
	if r.pcregexp != nil {
		return r.pcregexp.Match(b)
	}
	return r.regexp.Match(b)
}

func (r *Regexp) MatchString(s string) bool {
	if r.pcregexp != nil {
		return r.pcregexp.MatchString(s)
	}
	return r.regexp.MatchString(s)
}

func (r *Regexp) MatchReader(reader io.RuneReader) bool {
	if r.pcregexp != nil {
		return r.pcregexp.MatchReader(reader)
	}
	return r.regexp.MatchReader(reader)
}

func (r *Regexp) ReplaceAll(src, repl []byte) []byte {
	if r.pcregexp != nil {
		return r.pcregexp.ReplaceAll(src, repl)
	}
	return r.regexp.ReplaceAll(src, repl)
}

func (r *Regexp) ReplaceAllString(src, repl string) string {
	if r.pcregexp != nil {
		return r.pcregexp.ReplaceAllString(src, repl)
	}
	return r.regexp.ReplaceAllString(src, repl)
}

func (r *Regexp) ReplaceAllLiteral(src, repl []byte) []byte {
	if r.pcregexp != nil {
		return r.pcregexp.ReplaceAllLiteral(src, repl)
	}
	return r.regexp.ReplaceAllLiteral(src, repl)
}

func (r *Regexp) ReplaceAllLiteralString(src, repl string) string {
	if r.pcregexp != nil {
		return r.pcregexp.ReplaceAllLiteralString(src, repl)
	}
	return r.regexp.ReplaceAllLiteralString(src, repl)
}

func (r *Regexp) Expand(dst []byte, template []byte, src []byte, match []int) []byte {
	if r.pcregexp != nil {
		return r.pcregexp.Expand(dst, template, src, match)
	}
	return r.regexp.Expand(dst, template, src, match)
}

func (r *Regexp) ExpandString(dst []byte, template string, src string, match []int) []byte {
	if r.pcregexp != nil {
		return r.pcregexp.ExpandString(dst, template, src, match)
	}
	return r.regexp.ExpandString(dst, template, src, match)
}

func (r *Regexp) NumSubexp() int {
	if r.pcregexp != nil {
		return r.pcregexp.NumSubexp()
	}
	return r.regexp.NumSubexp()
}

func (r *Regexp) SubexpNames() []string {
	if r.pcregexp != nil {
		return r.pcregexp.SubexpNames()
	}
	return r.regexp.SubexpNames()
}

func (r *Regexp) SubexpIndex(name string) int {
	if r.pcregexp != nil {
		return r.pcregexp.SubexpIndex(name)
	}
	return r.regexp.SubexpIndex(name)
}

func (r *Regexp) Split(s string, n int) []string {
	if r.pcregexp != nil {
		return r.pcregexp.Split(s, n)
	}
	return r.regexp.Split(s, n)
}

func (r *Regexp) String() string {
	return r.pattern
}

func (r *Regexp) LiteralPrefix() (prefix string, complete bool) {
	if r.pcregexp != nil {
		return r.pcregexp.LiteralPrefix()
	}
	return r.regexp.LiteralPrefix()
}

func (r *Regexp) ReplaceAllFunc(src []byte, repl func([]byte) []byte) []byte {
	if r.pcregexp != nil {
		return r.pcregexp.ReplaceAllFunc(src, repl)
	}
	return r.regexp.ReplaceAllFunc(src, repl)
}

func (r *Regexp) ReplaceAllStringFunc(src string, repl func(string) string) string {
	if r.pcregexp != nil {
		return r.pcregexp.ReplaceAllStringFunc(src, repl)
	}
	return r.regexp.ReplaceAllStringFunc(src, repl)
}

func (r *Regexp) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *Regexp) UnmarshalText(text []byte) error {
	re, err := Compile(string(text))
	if err != nil {
		return err
	}
	*r = *re
	return nil
}

func (r *Regexp) Longest() {
	if r.pcregexp != nil {
		r.pcregexp.Longest()
		return
	}
	r.regexp.Longest()
}

func (r *Regexp) FindReaderIndex(reader io.RuneReader) []int {
	if r.pcregexp != nil {
		return r.pcregexp.FindReaderIndex(reader)
	}
	return r.regexp.FindReaderIndex(reader)
}

func (r *Regexp) FindReaderSubmatchIndex(reader io.RuneReader) []int {
	if r.pcregexp != nil {
		return r.pcregexp.FindReaderSubmatchIndex(reader)
	}
	return r.regexp.FindReaderSubmatchIndex(reader)
}

func (r *Regexp) FindSubmatchIndex(b []byte) []int {
	if r.pcregexp != nil {
		return r.pcregexp.FindSubmatchIndex(b)
	}
	return r.regexp.FindSubmatchIndex(b)
}

func (r *Regexp) FindStringSubmatchIndex(s string) []int {
	if r.pcregexp != nil {
		return r.pcregexp.FindStringSubmatchIndex(s)
	}
	return r.regexp.FindStringSubmatchIndex(s)
}

func (r *Regexp) FindAllSubmatchIndex(b []byte, n int) [][]int {
	if r.pcregexp != nil {
		return r.pcregexp.FindAllSubmatchIndex(b, n)
	}
	return r.regexp.FindAllSubmatchIndex(b, n)
}

func (r *Regexp) FindAllStringSubmatchIndex(s string, n int) [][]int {
	if r.pcregexp != nil {
		return r.pcregexp.FindAllStringSubmatchIndex(s, n)
	}
	return r.regexp.FindAllStringSubmatchIndex(s, n)
}
