package gomocker

import (
	"io"

	"github.com/jauhararifin/gotype"
)

const gomockerPath = "github.com/jauhararifin/gomocker"

var defaultMockerGenerator = &mockerGenerator{
	funcMockerGenerator: &funcMockerGenerator{
		namer: &defaultFuncMockerNamer{},
	},
	interfaceMockerGenerator: &interfaceMockerGenerator{
		funcMockerNamer:      &defaultFuncMockerNamer{},
		interfaceMockerNamer: &defaultInterfaceMockerNamer{},
	},
}

func GenerateMocker(typeSpecs []gotype.TypeSpec, w io.Writer, options ...GenerateMockerOption) error {
	return defaultMockerGenerator.GenerateMocker(typeSpecs, w, options...)
}
