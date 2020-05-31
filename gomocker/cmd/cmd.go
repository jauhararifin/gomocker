package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jauhararifin/gomocker"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "gomocker",
	Short:   "GoMocker is Golang mocker generator",
	Long:    "GoMocker generates golang structs to help you mock a function or an interface.",
}

var genCmd = &cobra.Command{
	Use:          "gen identifier...",
	Short:        "Generate mocker",
	SilenceUsage: true,
	RunE:         executeGen,
}

func init() {
	genCmd.Flags().String("source", "", "Your golang source code file. Can omit this when using go generate")
	genCmd.Flags().String("package", "", "The output mocker package name. By default it will be the same as the source")
	genCmd.Flags().String("output", "", "Generated mocker output filename, by default it's <source_file>_mock_gen.go")
	genCmd.Flags().Bool("force", false, "Force create the file if it is already exist")

	rootCmd.AddCommand(genCmd)
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func executeGen(cmd *cobra.Command, args []string) error {
	sourceFile, err := parseSourceFile(cmd)
	if err != nil {
		return err
	}

	outputFile, err := parseOutputFile(cmd, sourceFile)
	if err != nil {
		return err
	}

	pkgName, err := parseOutputPkg(cmd, outputFile)
	if err != nil {
		return err
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("please provide at least one function or interface you want to mock")
	}

	typeSpecs := make([]gomocker.TypeSpec, 0, len(args))
	for _, arg := range args {
		// TODO (jauhar.arifin): use current package if pargs[0] is missing
		parts := strings.Split(arg, ":")
		if len(parts) != 2 {
			return fmt.Errorf("please provide a valid function or interface name. The format is <package-path>:<name1>,<name2>,...")
		}

		names := strings.Split(parts[1], ",")
		for _, name := range names {
			typeSpecs = append(typeSpecs, gomocker.TypeSpec{
				PackagePath: parts[0],
				Name:        name,
			})
		}
	}

	if _, err := os.Stat(outputFile); err == nil && !force {
		return fmt.Errorf("file %s already exists", outputFile)
	}

	tempOutput := &bytes.Buffer{}

	file, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := gomocker.GenerateMocker(
		typeSpecs,
		tempOutput,
		gomocker.WithOutputPackagePath(pkgName),
		gomocker.WithInputFileName(path.Base(sourceFile)),
	); err != nil {
		return err
	}

	outputF, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	if _, err := outputF.Write(tempOutput.Bytes()); err != nil {
		return err
	}

	return nil
}

func parseSourceFile(cmd *cobra.Command) (string, error) {
	sourceFile, err := cmd.Flags().GetString("source")
	if err != nil {
		return "", err
	}
	if sourceFile == "" {
		sourceFile = os.Getenv("GOFILE")
	}
	if sourceFile == "" {
		return "", fmt.Errorf("please provide source file")
	}
	if path.Ext(sourceFile) != ".go" {
		return "", fmt.Errorf("please provide .go file as the source file")
	}
	return sourceFile, nil
}

func parseOutputPkg(cmd *cobra.Command, outputFile string) (string, error) {
	// TODO (jauhar.arifin): parse package name based on the output dir
	pkgName, err := cmd.Flags().GetString("package")
	if err != nil {
		return "", err
	}
	if pkgName != "" {
		return pkgName, nil
	}

	pkg := filepath.Base(filepath.Dir(outputFile))
	if pkg != "." {
		return pkg, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Base(wd), nil
}

func parseOutputFile(cmd *cobra.Command, sourceFile string) (string, error) {
	outputFile, err := cmd.Flags().GetString("output")
	if err != nil {
		return "", err
	}
	if outputFile == "" {
		outputFile = path.Join(
			path.Dir(sourceFile),
			strings.TrimSuffix(path.Base(sourceFile), path.Ext(sourceFile))+"_mock_gen.go",
		)
	}
	return outputFile, nil
}
