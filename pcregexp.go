package pcregexp

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
	"unicode/utf8"
	"unsafe"

	"github.com/ebitengine/purego"
)

func init() {
	var libPath string

	switch runtime.GOOS {
	case "darwin":
		libPath = "libpcre2-8.dylib"
	case "linux", "freebsd":
		libPath = "libpcre2-8.so"
	case "windows":
		libPath = "pcre2-8.dll"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}

	lib, err := openLibrary(libPath)
	if err != nil {
		panic(fmt.Errorf("failed to load %s: %w", libPath, err))
	}

	// Register the functions by their PCRE2 symbol names.
	// (For the 8-bit versions, the symbols are suffixed with "_8".)
	funcs := [][2]any{
		{&pcre2_compile, "pcre2_compile_8"},
		{&pcre2_code_free, "pcre2_code_free_8"},
		{&pcre2_pattern_info, "pcre2_pattern_info_8"},
		{&pcre2_match, "pcre2_match_8"},
		{&pcre2_match_data_create_from_pattern, "pcre2_match_data_create_from_pattern_8"},
		{&pcre2_match_data_free, "pcre2_match_data_free_8"},
		{&pcre2_get_ovector_pointer, "pcre2_get_ovector_pointer_8"},
		// Match context functions
		{&pcre2_match_context_create, "pcre2_match_context_create_8"},
		{&pcre2_match_context_free, "pcre2_match_context_free_8"},
		// {&pcre2_set_offset_limit, "pcre2_set_offset_limit_8"},
		// {&pcre2_set_heap_limit, "pcre2_set_heap_limit_8"},
		{&pcre2_set_match_limit, "pcre2_set_match_limit_8"},
		{&pcre2_set_recursion_limit, "pcre2_set_recursion_limit_8"},
	}

	for _, f := range funcs {
		purego.RegisterLibFunc(f[0], lib, f[1].(string))
	}

	runtime.SetFinalizer(globalFinalizerObject, func(_ *int) {
		if defaultMatchCtx != nil {
			pcre2_match_context_free(defaultMatchCtx.ptr)
			defaultMatchCtx = nil
		}
	})
}

// PCREgexp is a compiled regular expression.
type PCREgexp struct {
	pattern   string  // original pattern
	buf       []int   // cached match offsets
	code      uintptr // pointer to compiled pcre2_code
	matchData uintptr // cached match data
}

// Compile creates a new PCREgexp from pattern.
//
// Note: an empty pattern is considered valid but will match nothing. :shrug:
// In this case, calling [Close] is unnecessary.
func Compile(pattern string) (*PCREgexp, error) {
	var patPtr *uint8
	var errcode int32
	var errOffset uint64

	if len(pattern) == 0 {
		return &PCREgexp{code: 0, pattern: pattern}, nil
	}

	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&pattern))
	patPtr = (*uint8)(unsafe.Pointer(strHeader.Data))
	// patPtr = (*uint8)(unsafe.StringData(pattern))

	code := pcre2_compile(patPtr, uint64(len(pattern)), 0, &errcode, &errOffset, 0)
	if code == 0 {
		return nil, fmt.Errorf("pcre2_compile failed at offset %d, error code %d", errOffset, errcode)
	}

	return &PCREgexp{code: code, pattern: pattern}, nil
}

// MustCompile is like [Compile] but panics on error.
//
// Note: an empty pattern is considered valid but will match nothing. :shrug:
// In this case, calling [Close] is unnecessary.
func MustCompile(pattern string) *PCREgexp {
	re, err := Compile(pattern)
	if err != nil {
		panic(err)
	}

	return re
}

// Match reports whether the byte slice b
// contains any match of the regular expression pattern.
// More complicated queries need to use [Compile] and the full [Regexp] interface.
//
// Note: an empty pattern is considered valid but will match nothing. :shrug:
func Match(pattern string, b []byte) (matched bool, err error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	defer re.Close()
	return re.Match(b), nil
}

// MatchReader reports whether the text returned by the [io.RuneReader]
// contains any match of the regular expression pattern.
// More complicated queries need to use [Compile] and the full [Regexp] interface.
//
// Note: an empty pattern is considered valid but will match nothing. :shrug:
func MatchReader(pattern string, r io.RuneReader) (matched bool, err error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	defer re.Close()
	return re.MatchReader(r), nil
}

// MatchString reports whether the string s
// contains any match of the regular expression pattern.
// More complicated queries need to use [Compile] and the full [Regexp] interface.
//
// Note: an empty pattern is considered valid but will match nothing. :shrug:
func MatchString(pattern string, s string) (matched bool, err error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	defer re.Close()
	return re.MatchString(s), nil
}

// Close frees the resources associated with the compiled pattern.
func (re *PCREgexp) Close() {
	if re.matchData != 0 {
		pcre2_match_data_free(re.matchData)
		re.matchData = 0
	}

	if re.code != 0 {
		pcre2_code_free(re.code)
		re.code = 0
	}
}

// saveMatchData creates a new match data object if it doesn't exist yet.
//
// It returns the pointer to the match data object. The match data object is
// used to store the results of a match.
func (re *PCREgexp) saveMatchData() uintptr {
	if re.matchData == 0 {
		re.matchData = pcre2_match_data_create_from_pattern(re.code, 0)
	}

	return re.matchData
}

// match performs a PCRE2 match on the given subject.
//
// It returns a slice of start/end indexes as returned by PCRE2.
func (re *PCREgexp) match(subject []byte) []int {
	if re.code == 0 || len(subject) == 0 {
		return nil
	}

	md := re.saveMatchData()
	if md == 0 {
		return nil
	}

	var subjectPtr *uint8

	if len(subject) > 0 {
		subjectPtr = (*uint8)(ptr(&subject[0]))
	}

	// Use the default match context if available
	var matchCtxPtr uintptr
	if defaultMatchCtx != nil {
		matchCtxPtr = defaultMatchCtx.ptr
	}

	ret := pcre2_match(re.code, subjectPtr, uint64(len(subject)), 0, 0, md, matchCtxPtr)
	if ret < 0 {
		return nil
	}

	n := int(ret)
	reqLen := n * 2

	if cap(re.buf) < reqLen {
		re.buf = make([]int, reqLen)
	} else {
		re.buf = re.buf[:reqLen]
	}

	ovector := pcre2_get_ovector_pointer(md)
	if ovector == nil {
		return nil
	}

	size := unsafe.Sizeof(uint64(0))
	for i := 0; i < reqLen; i++ {
		ptr := (*uint64)(ptr(uintptr(ptr(ovector)) + uintptr(i)*size))
		re.buf[i] = int(*ptr)
	}

	return re.buf
}

// MatchString reports whether the Regexp matches the given string.
func (re *PCREgexp) MatchString(s string) bool {
	return re.match(stringToBytesUnsafe(s)) != nil
}

// FindString returns the text of the leftmost match in s.
func (re *PCREgexp) FindString(s string) string {
	indexes := re.match(stringToBytesUnsafe(s))
	if indexes == nil || len(indexes) < 2 {
		return ""
	}

	return s[indexes[0]:indexes[1]]
}

// FindStringIndex returns a two-element slice of integers defining the start
// and end of the leftmost match in s.
func (re *PCREgexp) FindStringIndex(s string) []int {
	return re.match(stringToBytesUnsafe(s))
}

// FindStringSubmatch returns a slice holding the text of the leftmost match and
// its submatches. It uses the actual number of captured groups as returned by
// PCRE2.
func (re *PCREgexp) FindStringSubmatch(s string) []string {
	indexes := re.match(stringToBytesUnsafe(s))
	if indexes == nil || len(indexes) < 2 {
		return nil
	}

	n := len(indexes) / 2
	submatches := make([]string, n)

	for i := 0; i < n; i++ {
		start := indexes[2*i]
		end := indexes[2*i+1]
		if start < 0 || end < 0 {
			submatches[i] = ""
		} else {
			submatches[i] = s[start:end]
		}
	}

	return submatches
}

// ReplaceAllString returns a copy of src in which all matches of the [PCREgexp]
// have been replaced by repl.
//
// If an empty match is encountered, it advances one UTF-8 rune to avoid
// infinite loop.
func (re *PCREgexp) ReplaceAllString(src, repl string) string {
	if src == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(src))

	remaining := src
	for {
		indexes := re.match(stringToBytesUnsafe(remaining))
		if indexes == nil || len(indexes) < 2 || indexes[0] < 0 {
			b.WriteString(remaining)
			break
		}

		b.WriteString(remaining[:indexes[0]])
		b.WriteString(repl)

		if indexes[0] == indexes[1] {
			if indexes[1] < len(remaining) {
				r, size := utf8.DecodeRuneInString(remaining[indexes[1]:])
				if r == utf8.RuneError || size == 0 {
					b.WriteString(remaining[indexes[1]:])
					break
				}

				b.WriteString(remaining[indexes[1] : indexes[1]+size])
				remaining = remaining[indexes[1]+size:]
			} else {
				remaining = ""
			}
		} else {
			remaining = remaining[indexes[1]:]
		}
	}

	return b.String()
}

// Find returns a slice holding the text of the leftmost match in b.
func (re *PCREgexp) Find(b []byte) []byte {
	indexes := re.match(b)
	if indexes == nil || len(indexes) < 2 {
		return nil
	}
	result := make([]byte, indexes[1]-indexes[0])
	copy(result, b[indexes[0]:indexes[1]])
	return result
}

// Match reports whether the regexp matches the byte slice b.
func (re *PCREgexp) Match(b []byte) bool {
	return re.match(b) != nil
}

// FindIndex returns a two-element slice of integers defining the location of
// the leftmost match in b.
func (re *PCREgexp) FindIndex(b []byte) []int {
	return re.match(b)
}

// FindSubmatch returns a slice of slices holding the text of the leftmost
// match and the matches of any subexpressions.
func (re *PCREgexp) FindSubmatch(b []byte) [][]byte {
	indexes := re.match(b)
	if indexes == nil || len(indexes) < 2 {
		return nil
	}

	matches := make([][]byte, len(indexes)/2)
	for i := 0; i < len(matches); i++ {
		start := indexes[2*i]
		end := indexes[2*i+1]
		if start < 0 || end < 0 {
			matches[i] = nil
		} else {
			matches[i] = make([]byte, end-start)
			copy(matches[i], b[start:end])
		}
	}
	return matches
}

// FindSubmatchIndex returns a slice holding the index pairs identifying the
// leftmost match and the matches of any subexpressions.
func (re *PCREgexp) FindSubmatchIndex(b []byte) []int {
	return re.match(b)
}

// FindReaderIndex returns a two-element slice of integers defining the location
// of the leftmost match in text read from the RuneReader. A return value of nil
// indicates no match.
func (re *PCREgexp) FindReaderIndex(r io.RuneReader) []int {
	var buf bytes.Buffer
	for {
		rune, _, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil
		}
		buf.WriteRune(rune)
	}

	// Get the match indices and return only the first pair
	if indexes := re.FindIndex(buf.Bytes()); len(indexes) >= 2 {
		return indexes[:2]
	}
	return nil
}

// FindReaderSubmatchIndex returns a slice holding the index pairs identifying
// the leftmost match of the regexp in text read from the RuneReader, and the
// matches of its subexpressions. A return value of nil indicates no match.
func (re *PCREgexp) FindReaderSubmatchIndex(r io.RuneReader) []int {
	// Convert RuneReader to []byte for processing
	var buf bytes.Buffer
	for {
		rune, _, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil
		}
		buf.WriteRune(rune)
	}
	return re.FindSubmatchIndex(buf.Bytes())
}

// MatchReader reports whether the regexp matches the text read from the
// RuneReader.
func (re *PCREgexp) MatchReader(r io.RuneReader) bool {
	// Convert RuneReader to []byte for processing
	var buf bytes.Buffer
	for {
		rune, _, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return false
		}
		buf.WriteRune(rune)
	}
	return re.Match(buf.Bytes())
}

// ReplaceAll returns a copy of src, replacing matches of the regexp with repl.
func (re *PCREgexp) ReplaceAll(src, repl []byte) []byte {
	return stringToBytesUnsafe(re.ReplaceAllString(string(src), string(repl)))
}

// NumSubexp returns the number of parenthesized subexpressions in this regexp.
//
// TODO(dwisiswant0): Implement this method.
func (re *PCREgexp) NumSubexp() int {
	// TODO(dwisiswant0): Implement this method.
	return 0
}

// String returns the source text used to compile the regexp.
func (re *PCREgexp) String() string {
	return re.pattern
}

// FindAllString returns a slice of all successive matches of the regexp in s.
// If n < 0, the return value contains all matches. If n >= 0, the return value
// contains at most n matches.
func (re *PCREgexp) FindAllString(s string, n int) []string {
	if n == 0 {
		return nil
	}

	var matches []string
	remaining := s

	for n != 0 {
		indexes := re.match(stringToBytesUnsafe(remaining))
		if indexes == nil || len(indexes) < 2 {
			break
		}

		matches = append(matches, remaining[indexes[0]:indexes[1]])

		if indexes[0] == indexes[1] {
			if indexes[1] >= len(remaining) {
				break
			}
			_, size := utf8.DecodeRuneInString(remaining[indexes[1]:])
			remaining = remaining[indexes[1]+size:]
		} else {
			remaining = remaining[indexes[1]:]
		}

		n--
	}

	return matches
}

// FindAllStringSubmatch is like [FindStringSubmatch] but returns successive
// matches.
func (re *PCREgexp) FindAllStringSubmatch(s string, n int) [][]string {
	if n == 0 {
		return nil
	}

	var results [][]string
	remaining := s

	for n != 0 {
		match := re.FindStringSubmatch(remaining)
		if match == nil {
			break
		}
		results = append(results, match)

		if len(match[0]) == 0 {
			if len(remaining) == 0 {
				break
			}
			_, size := utf8.DecodeRuneInString(remaining)
			remaining = remaining[size:]
		} else {
			remaining = remaining[len(match[0]):]
		}

		n--
	}

	return results
}

// FindAllStringIndex returns a slice of index pairs identifying successive
// matches of the regexp in s.
func (re *PCREgexp) FindAllStringIndex(s string, n int) [][]int {
	if n == 0 {
		return nil
	}

	var results [][]int
	remaining := s
	offset := 0

	for n != 0 {
		indexes := re.match(stringToBytesUnsafe(remaining))
		if indexes == nil || len(indexes) < 2 {
			break
		}

		// Adjust indexes for the offset
		adjIndexes := make([]int, 2)
		adjIndexes[0] = indexes[0] + offset
		adjIndexes[1] = indexes[1] + offset
		results = append(results, adjIndexes)

		if indexes[0] == indexes[1] {
			if indexes[1] >= len(remaining) {
				break
			}
			_, size := utf8.DecodeRuneInString(remaining[indexes[1]:])
			remaining = remaining[indexes[1]+size:]
			offset += indexes[1] + size
		} else {
			remaining = remaining[indexes[1]:]
			offset += indexes[1]
		}

		n--
	}

	return results
}

// ReplaceAllFunc returns a copy of src in which all matches of the regexp
// have been replaced by the return value of function repl applied to the
// matched byte slice.
func (re *PCREgexp) ReplaceAllFunc(src []byte, repl func([]byte) []byte) []byte {
	var b strings.Builder
	remaining := src
	lastMatchEnd := 0

	for {
		indexes := re.match(remaining)
		if indexes == nil || len(indexes) < 2 {
			b.Write(remaining[lastMatchEnd:])
			break
		}

		b.Write(remaining[:indexes[0]])
		match := remaining[indexes[0]:indexes[1]]
		b.Write(repl(match))

		if indexes[0] == indexes[1] {
			if indexes[1] >= len(remaining) {
				break
			}
			r, size := utf8.DecodeRune(remaining[indexes[1]:])
			if r == utf8.RuneError {
				b.Write(remaining[indexes[1]:])
				break
			}
			b.Write(remaining[indexes[1] : indexes[1]+size])
			remaining = remaining[indexes[1]+size:]
		} else {
			remaining = remaining[indexes[1]:]
		}
	}

	return stringToBytesUnsafe(b.String())
}

// Split slices s into substrings separated by matches of the regexp.
// If n > 0, Split returns at most n substrings, otherwise it returns all
// substrings.
func (re *PCREgexp) Split(s string, n int) []string {
	if n == 0 {
		return nil
	}

	var parts []string
	pos := 0
	for {
		indexes := re.FindStringIndex(s[pos:])
		if indexes == nil || indexes[0] < 0 {
			parts = append(parts, s[pos:])
			break
		}

		parts = append(parts, s[pos:pos+indexes[0]])
		pos += indexes[1]

		if n > 0 && len(parts) == n-1 {
			parts = append(parts, s[pos:])
			break
		}
	}
	return parts
}

// FindAll returns a slice of all successive matches of the regexp in b.
func (re *PCREgexp) FindAll(b []byte, n int) [][]byte {
	if n == 0 {
		return nil
	}

	var matches [][]byte
	remaining := b

	for n != 0 {
		indexes := re.match(remaining)
		if indexes == nil || len(indexes) < 2 {
			break
		}

		match := make([]byte, indexes[1]-indexes[0])
		copy(match, remaining[indexes[0]:indexes[1]])
		matches = append(matches, match)

		if indexes[0] == indexes[1] {
			if indexes[1] >= len(remaining) {
				break
			}
			_, size := utf8.DecodeRune(remaining[indexes[1]:])
			remaining = remaining[indexes[1]+size:]
		} else {
			remaining = remaining[indexes[1]:]
		}

		n--
	}

	return matches
}

// FindAllIndex returns a slice of index pairs identifying successive matches of
// the regexp in b.
func (re *PCREgexp) FindAllIndex(b []byte, n int) [][]int {
	if n == 0 {
		return nil
	}

	var results [][]int
	remaining := b
	offset := 0

	for n != 0 {
		indexes := re.match(remaining)
		if indexes == nil || len(indexes) < 2 {
			break
		}

		adjIndexes := make([]int, 2)
		adjIndexes[0] = indexes[0] + offset
		adjIndexes[1] = indexes[1] + offset
		results = append(results, adjIndexes)

		if indexes[0] == indexes[1] {
			if indexes[1] >= len(remaining) {
				break
			}
			_, size := utf8.DecodeRune(remaining[indexes[1]:])
			remaining = remaining[indexes[1]+size:]
			offset += indexes[1] + size
		} else {
			remaining = remaining[indexes[1]:]
			offset += indexes[1]
		}

		n--
	}

	return results
}

// ReplaceAllLiteral returns a copy of src, replacing matches of the regexp with
// repl.
func (re *PCREgexp) ReplaceAllLiteral(src, repl []byte) []byte {
	return re.ReplaceAll(src, repl)
}

// ReplaceAllLiteralString returns a copy of src, replacing matches of the
// regexp with repl.
func (re *PCREgexp) ReplaceAllLiteralString(src, repl string) string {
	return re.ReplaceAllString(src, repl)
}

// ReplaceAllStringFunc returns a copy of src in which all matches of the regexp
// have been replaced by the return value of function repl applied to the
// matched text.
func (re *PCREgexp) ReplaceAllStringFunc(src string, repl func(string) string) string {
	var b strings.Builder
	remaining := src
	lastMatchEnd := 0

	for {
		indexes := re.match(stringToBytesUnsafe(remaining))
		if indexes == nil || len(indexes) < 2 {
			b.WriteString(remaining[lastMatchEnd:])
			break
		}

		b.WriteString(remaining[:indexes[0]])
		match := remaining[indexes[0]:indexes[1]]
		b.WriteString(repl(match))

		if indexes[0] == indexes[1] {
			if indexes[1] >= len(remaining) {
				break
			}
			r, size := utf8.DecodeRuneInString(remaining[indexes[1]:])
			if r == utf8.RuneError {
				b.WriteString(remaining[indexes[1]:])
				break
			}
			b.WriteString(remaining[indexes[1] : indexes[1]+size])
			remaining = remaining[indexes[1]+size:]
		} else {
			remaining = remaining[indexes[1]:]
		}
	}

	return b.String()
}

// FindStringSubmatchIndex returns a slice holding the index pairs identifying
// the leftmost match of the regexp in s and the matches of its subexpressions.
func (re *PCREgexp) FindStringSubmatchIndex(s string) []int {
	return re.match(stringToBytesUnsafe(s))
}

// FindAllStringSubmatchIndex returns a slice of slices holding the index pairs
// identifying the successive matches of the regexp in s and their
// subexpressions.
func (re *PCREgexp) FindAllStringSubmatchIndex(s string, n int) [][]int {
	if n == 0 {
		return nil
	}

	var results [][]int
	remaining := s
	offset := 0

	for n != 0 {
		indexes := re.match(stringToBytesUnsafe(remaining))
		if indexes == nil || len(indexes) < 2 {
			break
		}

		// Adjust indexes for the offset
		adjIndexes := make([]int, len(indexes))
		for i := 0; i < len(indexes); i++ {
			if indexes[i] >= 0 {
				adjIndexes[i] = indexes[i] + offset
			} else {
				adjIndexes[i] = indexes[i]
			}
		}
		results = append(results, adjIndexes)

		if indexes[0] == indexes[1] {
			if indexes[1] >= len(remaining) {
				break
			}
			_, size := utf8.DecodeRuneInString(remaining[indexes[1]:])
			remaining = remaining[indexes[1]+size:]
			offset += indexes[1] + size
		} else {
			remaining = remaining[indexes[1]:]
			offset += indexes[1]
		}

		n--
	}

	return results
}

// FindAllSubmatch returns a slice of successive matches of the regexp in b.
func (re *PCREgexp) FindAllSubmatch(b []byte, n int) [][][]byte {
	if n == 0 {
		return nil
	}

	var results [][][]byte
	remaining := b

	for n != 0 {
		match := re.FindSubmatch(remaining)
		if match == nil {
			break
		}
		results = append(results, match)

		if len(match[0]) == 0 {
			if len(remaining) == 0 {
				break
			}
			_, size := utf8.DecodeRune(remaining)
			remaining = remaining[size:]
		} else {
			remaining = remaining[len(match[0]):]
		}

		n--
	}

	return results
}

// FindAllSubmatchIndex returns a slice of successive matches indexes of the
// regexp in b.
func (re *PCREgexp) FindAllSubmatchIndex(b []byte, n int) [][]int {
	if n == 0 {
		return nil
	}

	var results [][]int
	remaining := b
	offset := 0

	for n != 0 {
		indexes := re.match(remaining)
		if indexes == nil || len(indexes) < 2 {
			break
		}

		// Adjust indexes for the offset
		adjIndexes := make([]int, len(indexes))
		for i := 0; i < len(indexes); i++ {
			if indexes[i] >= 0 {
				adjIndexes[i] = indexes[i] + offset
			} else {
				adjIndexes[i] = indexes[i]
			}
		}
		results = append(results, adjIndexes)

		if indexes[0] == indexes[1] {
			if indexes[1] >= len(remaining) {
				break
			}
			_, size := utf8.DecodeRune(remaining[indexes[1]:])
			remaining = remaining[indexes[1]+size:]
			offset += indexes[1] + size
		} else {
			remaining = remaining[indexes[1]:]
			offset += indexes[1]
		}

		n--
	}

	return results
}

// Expand appends template to dst and returns the result; during the
// append, Expand replaces variables in the template with corresponding
// matches drawn from src.
func (re *PCREgexp) Expand(dst, template, src []byte, match []int) []byte {
	return re.expand(dst, string(template), src, match, false)
}

// ExpandString is like Expand but the template and source are strings.
func (re *PCREgexp) ExpandString(dst []byte, template, src string, match []int) []byte {
	return re.expand(dst, template, stringToBytesUnsafe(src), match, true)
}

// expand implements both Expand and ExpandString.
func (re *PCREgexp) expand(dst []byte, template string, src []byte, match []int, isString bool) []byte {
	for i := 0; i < len(template); i++ {
		if template[i] == '$' && i+1 < len(template) {
			switch template[i+1] {
			case '$':
				dst = append(dst, '$')
				i++
				continue
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				group := 0
				i++
				for i < len(template) && template[i] >= '0' && template[i] <= '9' {
					group = group*10 + int(template[i]-'0')
					i++
				}
				i--
				if 2*group < len(match) {
					dst = append(dst, src[match[2*group]:match[2*group+1]]...)
				}
				continue
			}
		}
		dst = append(dst, template[i])
	}
	return dst
}

// LiteralPrefix returns a literal string that must begin any match of the
// regular expression. It also returns a boolean indicating whether the literal
// is the entire regular expression.
//
// TODO(dwisiswant0): Implement this method.
func (re *PCREgexp) LiteralPrefix() (prefix string, complete bool) {
	// This would require PCRE2_INFO_PREFIX and PCRE2_INFO_PREFIXLENGTH
	// TODO(dwisiswant0): Implement using PCRE2 pattern info functions
	return "", false
}

// Longest makes future searches prefer the longest match.
// For PCRE2, this would require changing match flags, but since
// we're using a basic match function, this is currently a no-op.
//
// TODO(dwisiswant0): Implement this method.
func (re *PCREgexp) Longest() {
	// No-op for now
	// TODO(dwisiswant0): Implement using PCRE2 match options
}

// MarshalText implements the encoding.TextMarshaler interface.
func (re *PCREgexp) MarshalText() ([]byte, error) {
	return stringToBytesUnsafe(re.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (re *PCREgexp) UnmarshalText(text []byte) error {
	r, err := Compile(string(text))
	if err != nil {
		return err
	}
	*re = *r
	return nil
}

// SubexpNames returns the names of the parenthesized subexpressions
// in this regexp. The name for the first sub-expression is at index 1,
// following the same convention as index in FindSubmatch.
//
// TODO(dwisiswant0): Implement this method.
func (re *PCREgexp) SubexpNames() []string {
	// TODO(dwisiswant0): Implement using PCRE2_INFO_NAMECOUNT and
	// PCRE2_INFO_NAMETABLE
	return nil
}

// SubexpIndex returns the index of the first subexpression with the given name,
// or -1 if there is no subexpression with that name.
//
// TODO(dwisiswant0): Implement this method.
func (re *PCREgexp) SubexpIndex(name string) int {
	// TODO(dwisiswant0): Implement using PCRE2_INFO_NAMECOUNT and
	// PCRE2_INFO_NAMETABLE
	return -1
}
