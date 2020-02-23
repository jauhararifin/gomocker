package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"

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
	RunE:         executeGen,
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

func executeGen(cmd *cobra.Command, args []string) error {
	inputPackage, entityName, targetPackage, outputFile, err := parseParameters(cmd, args)
	if err != nil {
		return err
	}

	workingDir, f, err := prepareWorkingDir()
	if err != nil {
		return err
	}
	defer os.RemoveAll(workingDir)
	defer f.Close()

	if err := createMockGeneratorSource(
		inputPackage,
		entityName,
		targetPackage,
		outputFile,
		f,
	); err != nil {
		return err
	}

	outputBuff := &bytes.Buffer{}
	c := exec.Command("go", "run", "gomock_runner_temp/main.go")
	c.Stdout = outputBuff
	c.Stderr = outputBuff
	if err := c.Run(); err != nil {
		return fmt.Errorf("error when creating mocker: %s", outputBuff.String())
	}

	return nil
}

func parseParameters(
	cmd *cobra.Command,
	args []string,
) (inputPackage, entityName, targetPackage, outputFile string, err error) {
	if inputPackage, err = cmd.Flags().GetString("package"); err != nil {
		err = fmt.Errorf("cannot get input package name: %v", err)
		return
	}

	if targetPackage, err = getTargetPkg(cmd, inputPackage); err != nil {
		err = fmt.Errorf("cannot get target package name: %v", err)
		return
	}

	if outputFile, err = getOutputFile(cmd); err != nil {
		return
	}

	if len(args) == 0 {
		err = fmt.Errorf("please provide the function/interface names you wan't to mock")
		return
	}

	if len(args) != 1 {
		err = fmt.Errorf("please provide only one function/interface")
		return
	}

	entityName = args[0]
	return
}

func getTargetPkg(cmd *cobra.Command, inputPackage string) (string, error) {
	if cmd.Flag("target-package").Changed {
		return cmd.Flags().GetString("target-package")
	}
	return inputPackage, nil
}

func getOutputFile(cmd *cobra.Command) (string, error) {
	if cmd.Flag("output").Changed {
		return cmd.Flags().GetString("output")
	}

	outputFile := "mock.go"
	if o, ok := outputFileFromEnv(); ok {
		outputFile = o
	}

	_, err := os.Stat(outputFile)
	if os.IsNotExist(err) {
		return outputFile, nil
	} else if err != nil {
		return "", fmt.Errorf("cannot get output file: %w", err)
	}
	return "", fmt.Errorf("please provide --output flag")
}

func outputFileFromEnv() (string, bool) {
	gofile, ok := os.LookupEnv("GOFILE")
	if !ok {
		return "", false
	}

	if filepath.Ext(gofile) != ".go" {
		return "", false
	}

	base := filepath.Base(gofile)
	base = base[:len(base)-3]
	if base == "/" || base == "" {
		return "", false
	}

	if base[len(base)-1] == '_' {
		return base + "mock.go", ok
	}
	return base + "_mock.go", ok
}

func prepareWorkingDir() (directory string, f *os.File, err error) {
	dir := "gomock_runner_temp"
	if err = os.Mkdir(dir, 0777); err != nil {
		return "", nil, fmt.Errorf("cannot prepare directory to create generator: %v", err)
	}

	sourceFile := path.Join(dir, "main.go")
	if f, err = os.Create(sourceFile); err != nil {
		return dir, nil, fmt.Errorf("cannot create code generator")
	}

	return dir, f, nil
}

func createMockGeneratorSource(inputPackage, entityName, targetPackage, outputFile string, w io.Writer) error {
	code := jen.NewFile("main")
	code.Add(generateMockGenerator(inputPackage, entityName, targetPackage, outputFile))
	if err := code.Render(w); err != nil {
		return fmt.Errorf("cannot create code generator: %v", err)
	}
	return nil
}

func generateMockGenerator(inputPackage, entityName, targetPackage, outputFile string) jen.Code {
	return jen.Func().Id("main").Params().Block(
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
			jen.Qual("os", "Remove").Call(jen.Lit(outputFile)),
			jen.Panic(jen.Id("err")),
		),
	)
}
