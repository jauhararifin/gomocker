package main

import (
	"errors"
	"flag"
	"github.com/dave/jennifer/jen"
	"io"
	"log"
	"os"
	"os/exec"
)

const gomockerPath = "github.com/jauhararifin/gomocker"

func main() {
	var importPath, name, pkg, output string
	flag.StringVar(&importPath, "import", "", "The import path of your interface or function definition")
	flag.StringVar(&name, "name", "", "The name of your interface or function definition")
	flag.StringVar(&pkg, "package", "", "The name of your output package")
	flag.StringVar(&output, "output", "", "The output filename")
	flag.Parse()
	if err := run(importPath, name, pkg, output); err != nil {
		log.Fatalf("error=%v", err)
	}
}

func run(importPath, name, pkg, output string) error {
	if len(importPath) == 0 {
		return errors.New("missing import path")
	}
	if len(name) == 0 {
		return errors.New("missing interface/function name")
	}
	if len(pkg) == 0 {
		return errors.New("missing package name")
	}
	if len(output) == 0 {
		return errors.New("missing output target file")
	}

	if err := os.Mkdir("gomock_runner_temp", 0777); err != nil {
		return err
	}
	defer os.RemoveAll("gomock_runner_temp")

	generatorFile, err := os.Create("gomock_runner_temp/main.go")
	if err != nil {
		return err
	}
	defer generatorFile.Close()

	if err := generateCodeGenerator(importPath, name, pkg, output, generatorFile); err != nil {
		return err
	}

	return exec.Command("go", "run", "gomock_runner_temp/main.go").Run()
}

func generateCodeGenerator(importPath, name, pkg, output string, w io.Writer) error {
	code := jen.NewFile("main")
	code.Func().Id("main").Params().Block(
		jen.List(jen.Id("f"), jen.Id("err")).Op(":=").Qual("os", "Create").Call(jen.Lit(output)),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Panic(jen.Id("err")),
		),
		jen.Defer().Id("f").Dot("Close").Call(),
		jen.Id("err").Op("=").Qual(gomockerPath, "GenerateMocker").Call(
			jen.Qual("reflect", "TypeOf").
				Call(
					jen.Params(jen.Op("*").Qual(importPath, name)).Call(jen.Nil()),
				).Dot("Elem").Call(),
			jen.Qual(gomockerPath, "WithName").Call(jen.Lit(name)),
			jen.Qual(gomockerPath, "WithPackageName").Call(jen.Lit(pkg)),
			jen.Qual(gomockerPath, "WithWriter").Call(jen.Id("f")),
		),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Panic(jen.Id("err")),
		),
	)
	return code.Render(w)
}
