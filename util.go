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

	if s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		return parseIntHex(s[2:])
	} else if s[0] == '0' && (s[1] == 'b' || s[1] == 'B') {
		return parseIntBin(s[2:])
	} else if s[0] == '0' {
		if s[1] == 'o' || s[1] == 'O' {
			return parseIntOct(s[2:])
		}
		return parseIntOct(s[1:])
	}

	s = strings.ReplaceAll(s, "_", "")
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int(s[i]) - '0'
	}
	return n, true
}

func parseIntHex(s string) (int, bool) {
	s = strings.ReplaceAll(s, "_", "")
	n := 0
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] >= 'a' && s[i] <= 'f':
			n = n*16 + int(s[i]) - 'a' + 10
		case s[i] >= 'A' && s[i] <= 'F':
			n = n*16 + int(s[i]) - 'A' + 10
		case s[i] >= '0' && s[i] <= '9':
			n = n*16 + int(s[i]) - '0'
		}
	}
	return n, true
}

func parseIntOct(s string) (int, bool) {
	s = strings.ReplaceAll(s, "_", "")
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '7' {
			return 0, false
		}
		n = n*8 + int(s[i]) - '0'
	}
	return n, true
}

func parseIntBin(s string) (int, bool) {
	s = strings.ReplaceAll(s, "_", "")
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] != '0' && s[i] != '1' {
			return 0, false
		}
		n = n*2 + int(s[i]) - '0'
	}
	return n, true
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
