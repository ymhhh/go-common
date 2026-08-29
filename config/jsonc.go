package config

import "fmt"

// stripJSONComments removes // and /* */ comments from JSON-like content, while
// preserving string literals and escaped quotes. Comment bytes are replaced with
// whitespace so later line-oriented preprocessing cannot change column
// positions.
func stripJSONComments(in []byte) ([]byte, error) {
	out := make([]byte, 0, len(in))

	inStr := false
	escape := false
	for i := 0; i < len(in); i++ {
		c := in[i]

		if inStr {
			out = append(out, c)
			if escape {
				escape = false
				continue
			}
			if c == '\\' {
				escape = true
				continue
			}
			if c == '"' {
				inStr = false
			}
			continue
		}

		if c == '"' {
			inStr = true
			out = append(out, c)
			continue
		}

		if c != '/' || i+1 >= len(in) {
			out = append(out, c)
			continue
		}

		switch in[i+1] {
		case '/':
			out, i = stripLineComment(out, in, i)
		case '*':
			var err error
			out, i, err = stripBlockComment(out, in, i)
			if err != nil {
				return nil, err
			}
		default:
			out = append(out, c)
		}
	}

	return out, nil
}

// stripLineComment replaces a // comment starting at i with spaces.
// i must point at the first '/'. Returns the index of the last consumed byte.
func stripLineComment(out, in []byte, i int) ([]byte, int) {
	out = append(out, ' ', ' ')
	i += 2
	for i < len(in) && in[i] != '\n' {
		out = append(out, ' ')
		i++
	}
	if i < len(in) && in[i] == '\n' {
		out = append(out, '\n')
	}
	return out, i
}

// stripBlockComment replaces a /* */ comment starting at i with spaces,
// preserving newlines. i must point at the first '/'.
func stripBlockComment(out, in []byte, i int) ([]byte, int, error) {
	out = append(out, ' ', ' ')
	i += 2
	for i+1 < len(in) && (in[i] != '*' || in[i+1] != '/') {
		if in[i] == '\n' {
			out = append(out, '\n')
		} else {
			out = append(out, ' ')
		}
		i++
	}
	if i+1 >= len(in) {
		return nil, i, fmt.Errorf("config: unterminated JSON block comment")
	}
	out = append(out, ' ', ' ')
	i++ // skip '/'
	return out, i, nil
}
