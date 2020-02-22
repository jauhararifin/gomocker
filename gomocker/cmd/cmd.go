package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/dave/jennifer/jen"
	"github.com/spf13/cobra"
)

var gomockerPath = "github.com/jauhararifin/gomocker"

var rootCmd = &cobra.Command{
	Use:   "gomocker",
	Short: "GoMocker is Golang mocker generator",
	Long:  "GoMocker generates golang structs to help you mock a function or an interface.",
}

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate mocker",
	Run: func(cmd *cobra.Command, args []string) {
		inputPackage, err := cmd.Flags().GetString("package")
		if err != nil {
			log.Printf("cannot get input package name: %v\n", err)
			return
		}

		targetPackage, err := cmd.Flags().GetString("target-package")
		if err != nil {
			log.Printf("cannot get target package name: %v\n", err)
			return
		}

		outputFile, err := cmd.Flags().GetString("output")
		if err != nil {
			log.Printf("cannot get target output file: %v\n", err)
			return
		}

		if len(args) == 0 {
			log.Printf("please provide the function/interface names you wan't to mock\n")
			return
		}

		if len(args) != 1 {
			log.Printf("please provide only one function/interface\n")
			return
		}

		entityName := args[0]

		if err := os.Mkdir("gomock_runner_temp", 0777); err != nil {
			log.Printf("cannot prepare directory to create generator: %v\n", err)
			return
		}
		defer os.RemoveAll("gomock_runner_temp")

		generatorFile, err := os.Create("gomock_runner_temp/main.go")
		if err != nil {
			log.Printf("cannot create code generator\n")
			return
		}
		//defer generatorFile.Close()

		code := jen.NewFile("main")
		code.Func().Id("main").Params().Block(
			jen.List(jen.Id("f"), jen.Id("err")).Op(":=").Qual("os", "Create").Call(jen.Lit(outputFile)),
			jen.If(jen.Id("err").Op("!=").Nil()).Block(
				jen.Panic(jen.Id("err")),
			),
			jen.Defer().Id("f").Dot("Close").Call(),
			jen.Id("err").Op("=").Qual(gomockerPath, "GenerateMocker").Call(
				jen.Qual("reflect", "TypeOf").
					Call(
						jen.Params(jen.Op("*").Qual(inputPackage, entityName)).Call(jen.Nil()),
					).Dot("Elem").Call(),
				jen.Lit(entityName),
				jen.Lit(targetPackage),
				jen.Id("f"),
			),
			jen.If(jen.Id("err").Op("!=").Nil()).Block(
				jen.Panic(jen.Id("err")),
			),
		)

		if err := code.Render(generatorFile); err != nil {
			log.Printf("cannot create code generator: %v\n", err)
			return
		}

		outputBuff := &bytes.Buffer{}
		c := exec.Command("go", "run", "gomock_runner_temp/main.go")
		c.Stdout = outputBuff
		if err := c.Run(); err != nil {
			log.Printf("error when creating mocker: %s\n", outputBuff.String())
			return
		}
	},
}

func init() {
	genCmd.Flags().String("package", "", "Your function/interface package")
	genCmd.Flags().String("target-package", "mock", "Your target package name")
	genCmd.Flags().String("output", "mock.go", "Generated mocker output filename")

	rootCmd.AddCommand(genCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
