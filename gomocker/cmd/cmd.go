package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/jauhararifin/gomocker"
	"github.com/spf13/cobra"
)

var version = "v1.0.0"

var rootCmd = &cobra.Command{
	Use:   "gomocker",
	Short: "GoMocker is Golang mocker generator",
	Long:  "GoMocker generates golang structs to help you mock a function or an interface.",
}

var versionCmd = &cobra.Command{
	Use:          "version",
	Short:        "Print version",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
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
	rootCmd.AddCommand(versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func executeGen(cmd *cobra.Command, args []string) error {
	sourceFile, err := parseSourceFile(cmd)
	if err != nil {
		return err
	}

	pkgName, err := cmd.Flags().GetString("package")
	if err != nil {
		return err
	}

	outputFile, err := parseOutputFile(cmd, sourceFile)
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
		file,
		args,
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
