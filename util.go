package gomocker

import (
	"strings"

	"github.com/dave/jennifer/jen"
)

func makePublic(varname string) string {
	return strings.Title(varname)
}

type stepFunc func() jen.Code

func concatSteps(steps ...stepFunc) jen.Code {
	j := jen.Empty()
	for _, step := range steps {
		j.Add(step()).Line().Line()
	}
	return j
}
