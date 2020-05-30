package gomocker

import "io"

const gomockerPath = "github.com/jauhararifin/gomocker"

var defaultMockerGenerator = &mockerGenerator{
	astTypeGenerator: &astTypeGenerator{},
	funcMockerGenerator: &funcMockerGenerator{
		namer: &defaultFuncMockerNamer{},
	},
	interfaceMockerGenerator: &interfaceMockerGenerator{
		funcMockerNamer:      &defaultFuncMockerNamer{},
		interfaceMockerNamer: &defaultInterfaceMockerNamer{},
	},
}

func GenerateMocker(r io.Reader, names []string, w io.Writer, options ...GenerateMockerOption) error {
	return defaultMockerGenerator.GenerateMocker(r, names, w, options...)
}
