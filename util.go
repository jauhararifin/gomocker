package gomocker

import (
	"strings"

	"github.com/dave/jennifer/jen"
)

func parseInt(s string) (int, bool) {
	if len(s) == 0 {
		return 0, false
	}

	if len(s) == 1 {
		if s[0] < '0' || s[0] > '9' {
			return 0, false
		}
		return int(s[0]) - '0', true
	}

	for i := 1; i < len(s); i++ {
		if s[i] == '_' && s[i-1] == '_' {
			return 0, false
		}
	}
	if s[len(s)-1] == '_' || s[0] == '_' {
		return 0, false
	}

	if len(s) >= 3 && s[0] == '0' && s[1] == '_' && (s[2] < '0' || s[2] > '9') {
		return 0, false
	}

	base := uint8(10)
	if s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
		base = 16
	} else if s[0] == '0' && (s[1] == 'b' || s[1] == 'B') {
		s = s[2:]
		base = 2
	} else if s[0] == '0' {
		base = 8
		if s[1] == 'o' || s[1] == 'O' {
			s = s[2:]
		} else {
			s = s[1:]
		}
	}

	s = strings.ReplaceAll(s, "_", "")
	if len(s) == 0 {
		return 0, false
	}

	n := 0
	for i := 0; i < len(s); i++ {
		switch {
		case base >= 10 && s[i] >= 'a' && s[i] < 'a' + base - 10:
			n = n*int(base) + int(s[i]) - 'a' + 10
		case base >= 10 && s[i] >= 'A' && s[i] < 'A' + base - 10:
			n = n*int(base) + int(s[i]) - 'A' + 10
		case s[i] >= '0' && s[i] < '0'+maxUint8(base, 10):
			n = n*int(base) + int(s[i]) - '0'
		}
	}
	return n, true
}

func maxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

type stepFunc func() (jen.Code, error)

func concatSteps(steps ...stepFunc) (jen.Code, error) {
	j := jen.Empty()
	for _, step := range steps {
		code, err := step()
		if err != nil {
			return nil, err
		}
		j.Add(code).Line().Line()
	}
	return j, nil
}
