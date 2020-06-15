package gomocker

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func CheckPackageName(position string) (string, error) {
	absPath, err := filepath.Abs(position)
	if err != nil {
		return "", fmt.Errorf("cannot find absolute path: %w", err)
	}

	mod, goModPosition, err := FindModuleFile(absPath)
	if err != nil {
		return "", err
	}
	moduleName := mod.Module.Mod.Path
	goModPosition = path.Dir(goModPosition)

	return moduleName + absPath[len(goModPosition):], nil
}

func FindModuleFile(position string) (*modfile.File, string, error) {
	absPath, err := filepath.Abs(position)
	if err != nil {
		return nil, "", fmt.Errorf("cannot find absolute path: %w", err)
	}

	goModFilePath, err := FindGoModFile(absPath)
	if err != nil {
		return nil, "", fmt.Errorf("cannot find go.mod file: %w", err)
	}

	goModFile, err := os.Open(goModFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("cannot open go.mod file: %w", err)
	}
	defer goModFile.Close()

	goModBytes, err := ioutil.ReadAll(goModFile)
	if err != nil {
		return nil, "", fmt.Errorf("cannot read go.mod file: %w", err)
	}

	moduleFile, err := modfile.Parse("go.mod", goModBytes, nil)
	if err != nil {
		return nil, "", fmt.Errorf("cannot parse go.mod file: %w", err)
	}

	return moduleFile, goModFilePath, nil
}

// FindGoModFile find path to go.mod file from `position` folder
func FindGoModFile(position string) (string, error) {
	absPath, err := filepath.Abs(position)
	if err != nil {
		return "", fmt.Errorf("cannot find absolute path: %w", err)
	}

	maxDepth := 30
	for i := 0; i < maxDepth; i++ {
		goModPath := path.Join(absPath, "go.mod")
		if _, err := os.Stat(goModPath); errors.Is(err, os.ErrNotExist) {
			if absPath == "/" {
				break
			}
			absPath = path.Join(absPath, "..")
			continue
		}

		return goModPath, nil
	}

	return "", fmt.Errorf("no go.mod file found")
}
