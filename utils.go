package pcregexp

import (
	"strings"
	"unsafe"
)

// ptr aliases [unsafe.Pointer].
type ptr = unsafe.Pointer

// string2ByteSlice returns a byte slice header that points to the string's
// data. This conversion is safe only if the receiver does not modify the
// returned slice.
func string2BytesUnsafe(str string) []byte {
	if str == "" {
		return nil
	}

	return unsafe.Slice(unsafe.StringData(str), len(str))
}

// bytes2StringUnsafe converts a byte slice to a string. This conversion is
// safe only if the receiver does not modify the returned string.
func bytes2StringUnsafe(bs []byte) string {
	if len(bs) == 0 {
		return ""
	}

	return unsafe.String(unsafe.SliceData(bs), len(bs))
}

// NeedsPCRE checks if the pattern contains PCRE2-only features, based on
// pcre2syntax.
//
// Ref: https://pcre2project.github.io/pcre2/doc/pcre2syntax/
func NeedsPCRE(pattern string) bool {
	// Tokens and constructs
	pcre2Tokens := []string{
		// Lookahead/lookbehind assertions (atomic and non-atomic)
		"(?=", "(?!", "(?<=", "(?<!", // Perl-style lookarounds
		"(*pla:", "(*positive_lookahead:",
		"(*nla:", "(*negative_lookahead:",
		"(*plb:", "(*positive_lookbehind:",
		"(*nlb:", "(*negative_lookbehind:",
		"(?*", "(*napla:", "(*non_atomic_positive_lookahead:",
		"(?<*", "(*naplb:", "(*non_atomic_positive_lookbehind:",
		// Substring scan assertion
		"(*scan_substring:", "(*scs:",
		// Script runs
		"(*script_run:", "(*sr:", "(*atomic_script_run:", "(*asr:",
		// Backtracking control verbs
		"(*ACCEPT)", "(*FAIL)", "(*F)", "(*MARK:", "(*:", "(*COMMIT)", "(*PRUNE)", "(*SKIP)", "(*THEN)",
		// Option setting (PCRE2 extensions)
		"(*LIMIT_DEPTH=", "(*LIMIT_HEAP=", "(*LIMIT_MATCH=", "(*CASELESS_RESTRICT)", "(*NOTEMPTY)", "(*NOTEMPTY_ATSTART)",
		"(*NO_AUTO_POSSESS)", "(*NO_DOTSTAR_ANCHOR)", "(*NO_JIT)", "(*NO_START_OPT)", "(*TURKISH_CASING)", "(*UTF)", "(*UCP)",
		// Newline conventions
		"(*CR)", "(*LF)", "(*CRLF)", "(*ANYCRLF)", "(*ANY)", "(*NUL)",
		// What \R matches
		"(*BSR_ANYCRLF)", "(*BSR_UNICODE)",
		// Atomic groups
		"(?>", "(*atomic:",
		// Branch reset group
		"(?|",
		// Conditional group
		"(?(DEFINE)", "(?(", // (?(DEFINE) and (?(condition)
		// Comment
		"(?#",
		// Recursion/subroutine calls
		"(?R)", "(?P>", "(?&", // (?R), (?P>name), (?&name)
		// Perl extended character classes
		"(?[",
	}

	// Escapes and character types
	pcre2Escapes := []string{
		`(?C`,      // callout
		`\C`,       // one code unit (dangerous, not in Go)
		`\h`, `\H`, // horizontal whitespace
		`\v`, `\V`, // vertical whitespace
		`\R`,         // newline sequence
		`\X`,         // Unicode extended grapheme cluster
		`\N`,         // not newline (not supported in Go)
		`\K`,         // set reported start of match
		`\e`,         // escape character
		`\f`,         // form feed
		`\a`,         // alarm
		`\o{`,        // octal code
		`\N{U+`,      // Unicode code point
		`\x{`,        // hex code (Go only supports \xhh)
		`\p{`, `\P{`, // Unicode property escapes (Go supports a subset)
	}

	// Backreference and subroutine call syntax not supported by Go
	pcre2Backrefs := []string{
		`\\g`, `\\k`, // \g, \k (named/numbered backrefs)
		`(?P=`,         // Python-style named backref
		`\\g<`, `\\g'`, // Oniguruma-style subroutine call
	}

	// Named backreferences
	pcre2NamedBackrefs := []string{
		`\k<`, `\k'`, `\k{`,
	}

	// Named group call
	pcre2NamedGroupCall := []string{
		`(?P>`, `(?&`, // Perl-style named group call
	}

	// Anchors (Go supports ^ and $ only)
	pcre2Anchors := []string{`\A`, `\Z`, `\z`, `\G`}

	// Check for any PCRE2-only features
	checks := [][]string{
		pcre2Tokens,
		pcre2Escapes,
		pcre2Backrefs,
		pcre2NamedBackrefs,
		pcre2NamedGroupCall,
		pcre2Anchors,
	}
	for _, group := range checks {
		for _, v := range group {
			if strings.Contains(pattern, v) {
				return true
			}
		}
	}

	// Check for backreferences: \1, \2, ... (Go does not support these)
	escaped := false
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

	// Named capturing groups
	//
	// NOTE(dwisiswant0): This is a bit tricky. Go supports (?P<name>...), but
	// not (?'name'...) or (?<name>...)
	if !strings.Contains(pattern, "(?P<") &&
		(strings.Contains(pattern, "(?<") || strings.Contains(pattern, "(?'")) {
		return true
	}

	return false
}
