package gomocker

import "io"

const gomockerPath = "github.com/jauhararifin/gomocker"

var defaultMockerGenerator = &mockerGenerator{
	astTypeGenerator: &astTypeGenerator{
		sourceFinder: &defaultSourceFinder{},
	},
	funcMockerGenerator: &funcMockerGenerator{
		namer: &defaultFuncMockerNamer{},
	},
	interfaceMockerGenerator: &interfaceMockerGenerator{
		funcMockerNamer:      &defaultFuncMockerNamer{},
		interfaceMockerNamer: &defaultInterfaceMockerNamer{},
	},
}

func GenerateMocker(typeSpecs []TypeSpec, w io.Writer, options ...GenerateMockerOption) error {
	return defaultMockerGenerator.GenerateMocker(typeSpecs, w, options...)
}
