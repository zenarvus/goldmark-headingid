// package headingid is an extension for the goldmark (http://github.com/yuin/goldmark).
//
// This extension enhances the automatic heading ID generation.
package headingid

import (
	"bytes"
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// An IDs interface is a collection of the element ids.
type IDs interface {
	// Generate generates a new element id.
	Generate(value []byte, kind ast.NodeKind) []byte

	// Put puts a given element id to the used ids table.
	Put(value []byte)
}

type ids struct {
	values map[string]bool
}

func NewIDs() IDs {
	return &ids{
		values: map[string]bool{},
	}
}

func (s *ids) Generate(value []byte, kind ast.NodeKind) []byte {
	result := slugify(value,'-')
	if len(result) == 0 {
		if kind == ast.KindHeading {
			result = []byte("heading")
		} else {
			result = []byte("id")
		}
	}
	if _, ok := s.values[util.BytesToReadOnlyString(result)]; !ok {
		s.values[util.BytesToReadOnlyString(result)] = true
		return result
	}
	for i := 1; ; i++ {
		newResult := fmt.Sprintf("%s-%d", result, i)
		if _, ok := s.values[newResult]; !ok {
			s.values[newResult] = true
			return []byte(newResult)
		}
	}
}

func (s *ids) Put(value []byte) {
	s.values[util.BytesToReadOnlyString(value)] = true
}

// Map of common unicode runes to ASCII replacements.
var repl = map[rune]string{
	'á': "a", 'à': "a", 'â': "a", 'ä': "a", 'ã': "a", 'å': "a",
	'é': "e", 'è': "e", 'ê': "e", 'ë': "e",
	'í': "i", 'ì': "i", 'î': "i", 'ï': "i", 'ı': "i",
	'ó': "o", 'ò': "o", 'ô': "o", 'ö': "o", 'õ': "o",
	'ú': "u", 'ù': "u", 'û': "u", 'ü': "u",
	'ğ': "g", 'ñ': "n", 'ç': "c",
	'ý': "y", 'ÿ': "y",
	'þ': "th", 'ð': "d",
	'æ': "ae", 'œ': "oe",
}

// Slugify converts input bytes to a slug bytes slice using sep (e.g., '-').
// Result is lowercased ASCII; non-transliterable runes are removed or become sep.
func slugify(in []byte, sep byte) []byte {
	if len(in) == 0 {return nil}
	var out bytes.Buffer
	out.Grow(len(in))
	prevSep := false

	for len(in) > 0 {
		r, size := utf8.DecodeRune(in)
		in = in[size:]

		// ASCII fast-path
		if r < utf8.RuneSelf {
			switch {
			case r >= 'A' && r <= 'Z':
				out.WriteByte(byte(r + ('a' - 'A'))) // to lowercase
				prevSep = false
			case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
				out.WriteByte(byte(r))
				prevSep = false
			default:
				if !prevSep && out.Len() > 0 {
					out.WriteByte(sep)
					prevSep = true
				}
			}
			continue
		} else {
			// Transliterate common runes
			if s, ok := repl[unicode.ToLower(r)]; ok {
				// write transliteration as lowercase
				out.WriteString(s)
				prevSep = false
				continue
			}
		}

		// For letters in other scripts try to use unicode.IsLetter -> drop if not ASCII
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			// Attempt a naive decomposition: remove diacritics isn't cheap; drop unknown non-ASCII.
			// To keep short and fast, we omit expensive normalization and treat as separator.
			if !prevSep && out.Len() > 0 {
				out.WriteByte(sep)
				prevSep = true
			}
			continue
		}

		// Other runes -> separator
		if !prevSep && out.Len() > 0 {
			out.WriteByte(sep)
			prevSep = true
		}
	}
	// Trim trailing separator
	b := out.Bytes()
	if len(b) > 0 && b[len(b)-1] == sep { b = b[:len(b)-1] }
	// Trim leading separator
	if len(b) > 0 && b[0] == sep { b = b[1:] }
	// Return a copy to ensure external mutation won't affect internal buffer
	res := make([]byte, len(b))
	copy(res, b)
	return res
}
