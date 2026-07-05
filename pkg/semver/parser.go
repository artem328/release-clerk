package semver

import "errors"

var errInvalidSemVer = errors.New("invalid semver")

func Parse(s string) (Version, error) {
	return ParseBytes([]byte(s))
}

func ParseBytes(b []byte) (Version, error) {
	l := &lexer{data: b}

	var (
		v   Version
		err error
	)

	v.Major, err = parseCoreVersion(l)
	if err != nil {
		return Version{}, err
	}

	if next := l.next(); next != '.' {
		return Version{}, errInvalidSemVer
	}

	v.Minor, err = parseCoreVersion(l)
	if err != nil {
		return Version{}, err
	}

	if next := l.next(); next != '.' {
		return Version{}, errInvalidSemVer
	}

	v.Patch, err = parseCoreVersion(l)
	if err != nil {
		return Version{}, err
	}

	if l.peek() == '-' {
		l.next()
		v.PreRelease, err = parseIdentifiers(l)
		if err != nil {
			return Version{}, err
		}
	}

	if l.peek() == '+' {
		l.next()
		v.Build, err = parseIdentifiers(l)
		if err != nil {
			return Version{}, err
		}
	}

	if l.next() != 0 {
		return Version{}, errInvalidSemVer
	}

	return v, nil
}

func parseCoreVersion(l *lexer) (uint, error) {
	var v uint

	d := l.next()

	if !isDigit(d) {
		return 0, errInvalidSemVer
	}

	if d == '0' {
		if isDigit(l.peek()) {
			return 0, errInvalidSemVer
		}

		return 0, nil
	}

	v += uint(d - '0')

	for isDigit(l.peek()) {
		v *= 10
		v += uint(l.next() - '0')
	}

	return v, nil
}

func parseIdentifiers(l *lexer) ([]string, error) {
	var (
		p   []string
		buf = make([]byte, 0, 8)
	)

	for {
		n := l.peek()

		if n == '.' {
			l.next()
			if len(buf) == 0 {
				return nil, errInvalidSemVer
			}

			p = append(p, string(buf))
			buf = buf[:0]
			continue
		}

		if !isDigit(n) && !isLetter(n) && n != '-' {
			if len(buf) > 0 {
				p = append(p, string(buf))
			}

			break
		}

		l.next()

		buf = append(buf, n)
	}

	if len(p) == 0 {
		return nil, errInvalidSemVer
	}

	return p, nil
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isLetter(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

type lexer struct {
	data []byte
	pos  int
}

func (l *lexer) next() byte {
	if l.pos < len(l.data) {
		n := l.data[l.pos]
		l.pos++
		return n
	}

	return 0
}

func (l *lexer) peek() byte {
	if l.pos < len(l.data) {
		return l.data[l.pos]
	}

	return 0
}
