package gomocker

import "io"

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

// TypeSpec represents a combination of package path and the type's name which can uniquely identified Golang's type.
type TypeSpec struct {
	// PackagePath contains a defined type's package path, that is, the import path
	// that uniquely identifies the package, such as "encoding/base64".
	PackagePath string

	// Name contains the type's name inside the package.
	Name string
}

func GenerateMocker(typeSpecs []TypeSpec, w io.Writer, options ...GenerateMockerOption) error {
	return defaultMockerGenerator.GenerateMocker(typeSpecs, w, options...)
}
