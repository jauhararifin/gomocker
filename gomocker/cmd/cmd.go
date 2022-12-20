package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jauhararifin/gomocker"
	"github.com/jauhararifin/gotype"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gomocker",
	Short: "GoMocker is Golang mocker generator",
	Long:  "GoMocker generates golang structs to help you mock a function or an interface.",
}

var genCmd = &cobra.Command{
	Use:          "gen identifier...",
	Short:        "Generate mocker",
	SilenceUsage: true,
	RunE:         executeGen,
}

func Execute(version string) {
	rootCmd.Version = version
	// TODO (jauhar.arifin): actually we can infer the output package from the output dir.
	genCmd.Flags().String("package", "", "The output mocker package name. By default it will be the same as the source")
	genCmd.Flags().String("output", "", "Generated mocker output filename, by default it's <source_file>_mock_gen.go")
	genCmd.Flags().Bool("force", false, "Force create the file if it is already exist")
	rootCmd.AddCommand(genCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func executeGen(cmd *cobra.Command, args []string) error {
	sourceFile := os.Getenv("GOFILE")

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

	typeSpecs := make([]gotype.TypeSpec, 0, len(args))
	for _, arg := range args {
		packagePath, _ := parsePackageFromFilename(sourceFile)
		var names []string

		parts := strings.Split(arg, ":")
		if len(parts) == 1 {
			names = strings.Split(parts[0], ",")
		} else if len(parts) == 2 {
			packagePath = parts[0]
			names = strings.Split(parts[1], ",")
		} else {
			return fmt.Errorf("please provide a valid function or interface name. The format is <package-path>:<name1>,<name2>,...")
		}

		for _, name := range names {
			typeSpecs = append(typeSpecs, gotype.TypeSpec{
				PackagePath: packagePath,
				Name:        name,
			})
		}
	}

	if _, err := os.Stat(outputFile); err == nil && !force {
		return fmt.Errorf("file %s already exists", outputFile)
	}

	tempOutput := &bytes.Buffer{}

	if err := gomocker.GenerateMocker(
		typeSpecs,
		tempOutput,
		gomocker.WithOutputPackagePath(pkgName),
	); err != nil {
		return err
	}

	dir := path.Dir(outputFile)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("cannot create dir %s: %v", dir, err)
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

func parseOutputPkg(cmd *cobra.Command, outputFile string) (string, error) {
	// TODO (jauhar.arifin): parse package name based on the output dir
	pkgName, err := cmd.Flags().GetString("package")
	if err != nil {
		return "", err
	}
	if pkgName != "" {
		return pkgName, nil
	}

	return parsePackageFromFilename(outputFile)
}

func parsePackageFromFilename(filename string) (string, error) {
	pkgPathFromGoMod, err := gomocker.CheckPackageName(filepath.Dir(filename))
	if err == nil {
		return pkgPathFromGoMod, nil
	}

	pkg := filepath.Base(filepath.Dir(filename))
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

	if outputFile != "" {
		return outputFile, nil
	}

	if sourceFile == "" {
		return "mock.go", nil
	}

	sourceBase, sourceExt := path.Base(sourceFile), path.Ext(sourceFile)
	return path.Join(path.Dir(sourceFile), strings.TrimSuffix(sourceBase, sourceExt)+"_mock.go"), nil
}
