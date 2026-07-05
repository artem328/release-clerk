package conventionalcommit

import (
	"strings"
	"unicode"
)

const (
	footerTokenBreakingChange      = "BREAKING CHANGE"
	footerTokenBreakingChangeAlias = "BREAKING-CHANGE"
)

func Parse(raw string) Commit {
	if strings.TrimSpace(raw) == "" {
		return Commit{Raw: raw}
	}

	var c Commit

	c.Raw = raw

	lines := strings.Split(raw, "\n")

	c.RawHeader = lines[0]
	if !parseHeader(c.RawHeader, &c) {
		return Commit{Raw: raw}
	}

	if len(lines) > 1 && !parseBodyAndFooter(lines[1:], &c) {
		return Commit{Raw: raw}
	}

	c.IsConventional = true

	return c
}

func parseHeader(header string, c *Commit) bool {
	var ok bool

	data := []rune(header)

	c.Type, data, ok = parseCommitType(data)
	if !ok {
		return false
	}

	switch peek(data) {
	case '(':
		c.Scope, data, ok = parseScope(data)
		if !ok {
			return false
		}
	case 0:
		return false
	}

	switch peek(data) {
	case '!':
		c.IsBreaking = true
		data = cut(data, 1)
	case 0:
		return false
	}

	switch peek(data) {
	case ':':
		data = cut(data, 1)
		if peek(data) != ' ' {
			return false
		}
		data = cut(data, 1)
	default:
		return false
	}

	c.Description = strings.TrimSpace(string(data))
	if c.Description == "" {
		return false
	}

	return true
}

func parseCommitType(data []rune) (string, []rune, bool) {
	for i, r := range data {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) && r != '-' {
			if i == 0 {
				return "", data, false
			}

			return string(data[:i]), cut(data, i), true
		}
	}

	return string(data), nil, true
}

func parseScope(data []rune) (string, []rune, bool) {
	if peek(data) != '(' {
		return "", data, false
	}

	data = cut(data, 1)

	for i := 0; i < len(data); i++ {
		if data[i] == ')' {
			if i == 0 {
				return "", data, false
			}

			return string(data[:i]), cut(data, i+1), true
		}
	}

	return "", nil, false
}

func parseBodyAndFooter(lines []string, c *Commit) bool {
	if len(lines) == 0 {
		return true
	}

	if lines[0] != "" {
		return false
	}

	lines = cut(lines, 1)

	if len(lines) == 0 {
		return true
	}

	var (
		isEmptyPrevLine bool
		i               int
	)

	for i = 0; i < len(lines); i++ {
		if lines[i] == "" {
			isEmptyPrevLine = true
			continue
		}

		if (i == 0 || isEmptyPrevLine) && isFooterLine(lines[i]) {
			break
		}

		isEmptyPrevLine = false
	}

	c.Body = strings.TrimSpace(strings.Join(lines[:i], "\n"))
	lines = cut(lines, i)

	c.RawFooter = strings.TrimSpace(strings.Join(lines, "\n"))

	var (
		footer Footer
		value  []string
	)

	for i = 0; i < len(lines); i++ {
		tok, val, ok := extractFooterTokenAndValue(lines[i])
		if !ok {
			value = append(value, val)
			continue
		}

		if isBreakingChangeFooter(tok) {
			c.IsBreaking = true
		}

		if footer.Token != "" {
			footer.Value = strings.TrimSpace(strings.Join(value, "\n"))
			c.Footers = append(c.Footers, footer)
			value = value[:0]
		}

		footer.Token = tok
		value = append(value, val)
	}

	if footer.Token != "" {
		footer.Value = strings.TrimSpace(strings.Join(value, "\n"))
		c.Footers = append(c.Footers, footer)
	}

	return true
}

func isFooterLine(line string) bool {
	_, _, ok := extractFooterTokenAndValue(line)

	return ok
}

func extractFooterTokenAndValue(line string) (token string, value string, ok bool) {
	possibleFooter := strings.SplitN(line, ": ", 2)
	if len(possibleFooter) == 2 && isFooterToken(possibleFooter[0]) {
		return possibleFooter[0], possibleFooter[1], true
	}

	possibleFooter = strings.SplitN(line, " #", 2)
	if len(possibleFooter) == 2 && isFooterToken(possibleFooter[0]) {
		return possibleFooter[0], "#" + possibleFooter[1], true
	}

	return "", "", false
}

func isBreakingChangeFooter(token string) bool {
	return token == footerTokenBreakingChange || token == footerTokenBreakingChangeAlias
}

func isFooterToken(token string) bool {
	if isBreakingChangeFooter(token) {
		return true
	}

	for _, r := range token {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) && r != '-' {
			return false
		}
	}

	return true
}

func cut[T any](data []T, pos int) []T {
	if pos == 0 {
		return data
	}

	if pos >= len(data) {
		return nil
	}

	return data[pos:]
}

func peek(data []rune) rune {
	if len(data) == 0 {
		return 0
	}

	return data[0]
}
