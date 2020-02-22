package gomocker

import (
	"errors"
	"io"
	"reflect"

	"github.com/dave/jennifer/jen"
)

func GenerateMocker(t reflect.Type, name, packageName string, w io.Writer) (err error) {
	if t == nil {
		return errors.New("type is nil")
	}

	code := jen.Code(nil)
	if t.Kind() == reflect.Func {
		if code, err = GenerateFuncMocker(t, name); err != nil {
			return err
		}
	} else if t.Kind() == reflect.Interface {
		if code, err = GenerateServiceMocker(t); err != nil {
			return err
		}
	} else {
		return errors.New("unrecognized type")
	}

	file := jen.NewFilePath(packageName)
	file.Add(code)
	return file.Render(w)
}
