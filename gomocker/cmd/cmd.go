package cmd

import (
	"bytes"
	"fmt"
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
	Use:          "gen",
	Short:        "Generate mocker",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPackage, err := cmd.Flags().GetString("package")
		if err != nil {
			return fmt.Errorf("cannot get input package name: %v", err)
		}

		targetPackage, err := cmd.Flags().GetString("target-package")
		if err != nil {
			return fmt.Errorf("cannot get target package name: %v", err)
		}

		outputFile, err := cmd.Flags().GetString("output")
		if err != nil {
			return fmt.Errorf("cannot get target output file: %v", err)
		}

		if len(args) == 0 {
			return fmt.Errorf("please provide the function/interface names you wan't to mock")
		}

		if len(args) != 1 {
			return fmt.Errorf("please provide only one function/interface")
		}

		entityName := args[0]

		if err := os.Mkdir("gomock_runner_temp", 0777); err != nil {
			return fmt.Errorf("cannot prepare directory to create generator: %v", err)
		}
		defer os.RemoveAll("gomock_runner_temp")

		generatorFile, err := os.Create("gomock_runner_temp/main.go")
		if err != nil {
			return fmt.Errorf("cannot create code generator")
		}
		defer generatorFile.Close()

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
				jen.Id("f").Dot("Close").Call(),
				jen.Qual("os","Remove").Call(jen.Lit(outputFile)),
				jen.Panic(jen.Id("err")),
			),
		)

		if err := code.Render(generatorFile); err != nil {
			return fmt.Errorf("cannot create code generator: %v", err)
		}

		outputBuff := &bytes.Buffer{}
		c := exec.Command("go", "run", "gomock_runner_temp/main.go")
		c.Stdout = outputBuff
		c.Stderr = outputBuff
		if err := c.Run(); err != nil {
			fmt.Println("wow", err, outputBuff.String())
			return fmt.Errorf("error when creating mocker: %s", outputBuff.String())
		}

		return nil
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
