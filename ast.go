package gomocker

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

type astTypeGenerator struct{}

func (f *astTypeGenerator) GenerateTypesFromSpecs(typeSpecs ...TypeSpec) []Type {
	packagePathToSpecs := f.groupTypeSpecByPackage(typeSpecs)

	resultMap := make(map[TypeSpec]Type)
	for packagePath, specs := range packagePathToSpecs {
		for i, typ := range f.generateTypesInSinglePackage(packagePath, specs...) {
			resultMap[TypeSpec{PackagePath: packagePath, Name: specs[i]}] = typ
		}
	}

	results := make([]Type, 0, len(typeSpecs))
	for _, spec := range typeSpecs {
		results = append(results, resultMap[spec])
	}
	return results
}

func (f *astTypeGenerator) groupTypeSpecByPackage(typeSpecs []TypeSpec) map[string][]string {
	result := make(map[string][]string)
	for _, spec := range typeSpecs {
		result[spec.PackagePath] = append(result[spec.PackagePath], spec.Name)
	}
	return result
}

func (f *astTypeGenerator) generateTypesInSinglePackage(packagePath string, names ...string) []Type {
	packageDir := f.findGoPackageDir(packagePath)
	types := f.generateTypesFromDir(packageDir, packagePath, names...)
	results := make([]Type, 0, len(names))
	for _, typ := range types {
		results = append(results, typ)
	}
	return results
}

func (f *astTypeGenerator) findGoPackageDir(packagePath string) string {
	moduleFile, goModFilePath := f.findModuleFile()
	if packageDir, ok := f.findPackageDirByModulePath(
		packagePath,
		moduleFile.Module.Mod.Path,
		filepath.Dir(goModFilePath),
	); ok {
		return packageDir
	}

	for _, req := range moduleFile.Require {
		modPath := f.findPackagePathByVersion(req.Mod)
		if packageDir, ok := f.findPackageDirByModulePath(packagePath, req.Mod.Path, modPath); ok {
			return packageDir
		}
	}

	modPath := f.findPackagePathByVersion(module.Version{})
	if packageDir, ok := f.findPackageDirByModulePath(packagePath, "", modPath); ok {
		return packageDir
	}

	panic(fmt.Errorf("package %s cannot be found in go.mod file", packagePath))
}

func (f *astTypeGenerator) findModuleFile() (*modfile.File, string) {
	goModFilePath := f.findGoModFile()

	goModFile, err := os.Open(goModFilePath)
	if err != nil {
		panic(fmt.Errorf("cannot open go.mod file: %w", err))
	}
	defer goModFile.Close()

	goModBytes, err := ioutil.ReadAll(goModFile)
	if err != nil {
		panic(fmt.Errorf("cannot read go.mod file: %w", err))
	}

	moduleFile, err := modfile.Parse("go.mod", goModBytes, nil)
	if err != nil {
		panic(fmt.Errorf("cannot parse go.mod file: %w", err))
	}

	return moduleFile, goModFilePath
}

func (f *astTypeGenerator) findGoModFile() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("cannot get current working dir: %w", err))
	}

	maxDepth := 15
	for i := 0; i < maxDepth; i++ {
		goModPath := path.Join(wd, "go.mod")
		if _, err := os.Stat(goModPath); errors.Is(err, os.ErrNotExist) {
			if wd == "/" {
				break
			}
			wd = path.Join(wd, "..")
			continue
		}

		return goModPath
	}

	panic(fmt.Errorf("no go.mod file found"))
}

func (f *astTypeGenerator) findPackageDirByModulePath(
	targetPackagePath,
	modulePath,
	moduleDir string,
) (packageDir string, ok bool) {
	if !strings.HasPrefix(targetPackagePath, modulePath) {
		return "", false
	}
	return filepath.Join(moduleDir, targetPackagePath[len(modulePath):]), true
}

func (f *astTypeGenerator) findPackagePathByVersion(modVer module.Version) string {
	lookupDir := make([]string, 0, 0)

	if gohome, ok := os.LookupEnv("GOHOME"); ok {
		lookupDir = append(lookupDir, filepath.Join(gohome, "pkg", "mod", modVer.Path+"@"+modVer.Version))
	}

	// TODO (jauhar.arifin): this is for unix-based OS only
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("cannot get user home dir: %w", err))
	}
	lookupDir = append(lookupDir, filepath.Join(homedir, "go", "pkg", "mod", modVer.Path+"@"+modVer.Version))

	lookupDir = append(lookupDir, filepath.Join("/", "usr", "local", "go", "src"))

	if gohome, ok := os.LookupEnv("GOHOME"); ok {
		lookupDir = append(lookupDir, filepath.Join(gohome, "src", "mod", modVer.Path+"@"+modVer.Version))
	}

	for _, modulePath := range lookupDir {
		if dir, ok := f.findPackagePathFromCandidatePath(modulePath); ok {
			return dir
		}
	}

	panic(fmt.Errorf("cannot find module %s", modVer.String()))
}

func (f *astTypeGenerator) findPackagePathFromCandidatePath(modulePath string) (string, bool) {
	s, err := os.Stat(modulePath)
	if errors.Is(err, os.ErrNotExist) {
		return "", false
	}

	if !s.IsDir() {
		return "", false
	}
	return modulePath, true
}

func (f *astTypeGenerator) generateTypesFromDir(packageDir string, packagePath string, names ...string) []Type {
	goSources := f.getGoFilesInsideDir(packageDir)

	remainingNames := make(map[string]struct{})
	for _, name := range names {
		remainingNames[name] = struct{}{}
	}

	resultMap := make(map[string]Type)
	for _, source := range goSources {
		if len(remainingNames) == 0 {
			break
		}

		fileAst := f.parseAstFile(source)
		importMap := f.generateImportMap(fileAst)

		for name := range remainingNames {
			spec := f.getDeclarationByName(fileAst, name)
			if spec != nil {
				resultMap[name] = f.generateTypeFromExpr(spec.Type, packagePath, importMap)
				delete(remainingNames, name)
			}
		}
	}

	if len(remainingNames) != 0 {
		// TODO (jauhararifin): give better error message
		for name := range remainingNames {
			panic(fmt.Errorf("cannot find definition of %s", name))
		}
	}

	results := make([]Type, 0, len(names))
	for _, name := range names {
		results = append(results, resultMap[name])
	}

	return results
}

func (f *astTypeGenerator) getGoFilesInsideDir(dir string) []string {
	goSources := make([]string, 0, 0)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if path == dir {
			return nil
		}

		if err != nil {
			panic(fmt.Errorf("got error when listing file: %w", err))
		}

		if info.IsDir() {
			return filepath.SkipDir
		}

		if filepath.Ext(info.Name()) == ".go" {
			goSources = append(goSources, path)
		}

		return nil
	})
	return goSources
}

func (f *astTypeGenerator) parseAstFile(filename string) *ast.File {
	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Errorf("cannot open file: %w", err))
	}
	defer file.Close()

	fset := token.NewFileSet()
	fileAst, err := parser.ParseFile(fset, filepath.Base(filename), file, 0)
	if err != nil {
		panic(fmt.Errorf("cannot parse go code: %w", err))
	}
	return fileAst
}

func (f *astTypeGenerator) generateImportMap(file *ast.File) map[string]string {
	importMap := make(map[string]string)
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					importPath := importSpec.Path.Value
					importPath = strings.TrimPrefix(strings.TrimSuffix(importPath, "\""), "\"")

					importName := f.getImportNameFromPackagePath(importPath)
					if importSpec.Name != nil {
						importName = importSpec.Name.String()
					}

					importMap[importName] = importPath
				}
			}
		}
	}
	return importMap
}

func (f *astTypeGenerator) getImportNameFromPackagePath(packagePath string) string {
	// TODO (jauhararifin): check the correctness of this
	base := path.Base(packagePath)
	if base == "" {
		return ""
	}
	return strings.Split(base, ".")[0]
}

func (*astTypeGenerator) getDeclarationByName(f *ast.File, name string) *ast.TypeSpec {
	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if typeSpec.Name.String() == name {
						_, IsInterface := typeSpec.Type.(*ast.InterfaceType)
						_, IsFunction := typeSpec.Type.(*ast.FuncType)
						if IsInterface || IsFunction {
							return typeSpec
						}
					}
				}
			}
		}
	}
	return nil
}

func (f *astTypeGenerator) generateTypeFromExpr(e ast.Expr, targetPkgPath string, importMap map[string]string) Type {
	switch v := e.(type) {
	case *ast.SelectorExpr:
		return f.generateTypeFromSelectorExpr(v, importMap)
	case *ast.Ident:
		return f.generateTypeFromIdent(v, targetPkgPath)
	case *ast.StarExpr:
		typ := f.generateTypeFromStarExpr(v, targetPkgPath, importMap)
		return Type{PtrType: &typ}
	case *ast.ArrayType:
		return f.generateTypeFromArrayType(v, targetPkgPath, importMap)
	case *ast.FuncType:
		typ := f.generateTypeFromFuncType(v, targetPkgPath, importMap)
		return Type{FuncType: &typ}
	case *ast.MapType:
		typ := f.generateTypeFromMapType(v, targetPkgPath, importMap)
		return Type{MapType: &typ}
	case *ast.ChanType:
		typ := f.generateTypeFromChanType(v, targetPkgPath, importMap)
		return Type{ChanType: &typ}
	case *ast.StructType:
		typ := f.generateTypeFromStructType(v, targetPkgPath, importMap)
		return Type{StructType: &typ}
	case *ast.InterfaceType:
		typ := f.generateTypeFromInterfaceType(v, targetPkgPath, importMap)
		return Type{InterfaceType: &typ}
	}
	panic(fmt.Errorf("unrecognized type: %v", e))
}

func (f *astTypeGenerator) generateTypeFromIdent(ident *ast.Ident, targetPackagePath string) Type {
	switch ident.Name {
	case string(PrimitiveKindBool):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindBool}}
	case string(PrimitiveKindByte):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindByte}}
	case string(PrimitiveKindInt):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindInt}}
	case string(PrimitiveKindInt8):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindInt8}}
	case string(PrimitiveKindInt16):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindInt16}}
	case string(PrimitiveKindInt32):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindInt32}}
	case string(PrimitiveKindInt64):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindInt64}}
	case string(PrimitiveKindUint):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindUint}}
	case string(PrimitiveKindUint8):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindUint8}}
	case string(PrimitiveKindUint16):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindUint16}}
	case string(PrimitiveKindUint32):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindUint32}}
	case string(PrimitiveKindUint64):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindUint64}}
	case string(PrimitiveKindUintptr):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindUintptr}}
	case string(PrimitiveKindFloat32):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindFloat32}}
	case string(PrimitiveKindFloat64):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindFloat64}}
	case string(PrimitiveKindComplex64):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindComplex64}}
	case string(PrimitiveKindComplex128):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindComplex128}}
	case string(PrimitiveKindString):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindString}}
	case "error":
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindError}}
	}
	return Type{QualType: &QualType{Package: targetPackagePath, Name: ident.Name}}
}

func (f *astTypeGenerator) generateTypeFromSelectorExpr(
	selectorExpr *ast.SelectorExpr,
	importMap map[string]string,
) Type {
	ident, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		panic(fmt.Errorf("unrecognized expr: %v", selectorExpr))
	}

	importPath, ok := importMap[ident.String()]
	if !ok {
		panic(fmt.Errorf("unrecognized identifier: %s", ident.String()))
	}
	return Type{QualType: &QualType{
		Package: importPath,
		Name:    selectorExpr.Sel.String(),
	}}
}

func (f *astTypeGenerator) generateTypeFromStarExpr(
	starExpr *ast.StarExpr,
	targetPackagePath string,
	importMap map[string]string,
) PtrType {
	return PtrType{Elem: f.generateTypeFromExpr(starExpr.X, targetPackagePath, importMap)}
}

func (f *astTypeGenerator) generateTypeFromArrayType(
	arrayType *ast.ArrayType,
	targetPackagePath string,
	importMap map[string]string,
) Type {
	if arrayType.Len == nil {
		return Type{SliceType: &SliceType{Elem: f.generateTypeFromExpr(arrayType.Elt, targetPackagePath, importMap)}}
	}

	lit, ok := arrayType.Len.(*ast.BasicLit)
	if !ok {
		panic(fmt.Errorf("unrecognized array length: %v", arrayType.Len))
	}
	lenn, ok := parseInt(lit.Value)
	if !ok {
		panic(fmt.Errorf("unrecognized array length: %v", lit.Value))
	}

	return Type{ArrayType: &ArrayType{
		Len:  lenn,
		Elem: f.generateTypeFromExpr(arrayType.Elt, targetPackagePath, importMap),
	}}
}

func (f *astTypeGenerator) generateTypeFromFuncType(funcType *ast.FuncType, targetPackagePath string, importMap map[string]string) FuncType {
	params, isVariadic := f.generateTypeFromFieldList(
		funcType.Params,
		f.getInputNamesFromAst(funcType.Params.List),
		targetPackagePath,
		importMap,
	)

	var results []TypeField = nil
	if funcType.Results != nil {
		results, _ = f.generateTypeFromFieldList(
			funcType.Results,
			f.getOutputNamesFromAst(funcType.Results.List),
			targetPackagePath,
			importMap,
		)
	}
	return FuncType{
		Inputs:     params,
		Outputs:    results,
		IsVariadic: isVariadic,
	}
}

func (f *astTypeGenerator) generateTypeFromFieldList(
	fields *ast.FieldList,
	names []string,
	targetPackagePath string,
	importMap map[string]string,
) (types []TypeField, isVariadic bool) {
	if fields == nil {
		return nil, false
	}

	types = make([]TypeField, 0, fields.NumFields())
	i := 0
	for _, field := range fields.List {
		typExpr := field.Type
		if v, ok := field.Type.(*ast.Ellipsis); ok {
			isVariadic = true
			typExpr = v.Elt
		}
		typ := f.generateTypeFromExpr(typExpr, targetPackagePath, importMap)

		if len(field.Names) == 0 {
			types = append(types, TypeField{
				Name: names[i],
				Type: typ,
			})
			i++
			continue
		}

		for range field.Names {
			types = append(types, TypeField{
				Name: names[i],
				Type: typ,
			})
			i++
		}
	}

	return
}

func (f *astTypeGenerator) getInputNamesFromAst(inputs []*ast.Field) []string {
	return f.getNamesFromExpr(inputs, "arg")
}

func (f *astTypeGenerator) getOutputNamesFromAst(inputs []*ast.Field) []string {
	return f.getNamesFromExpr(inputs, "out")
}

func (f *astTypeGenerator) getNamesFromExpr(params []*ast.Field, prefix string) []string {
	names := make([]string, 0, len(params))
	i := 0
	for _, p := range params {
		n := len(p.Names)
		if n == 0 {
			n = 1
		}
		for j := 0; j < n; j++ {
			if len(p.Names) > 0 {
				names = append(names, p.Names[j].String())
			} else {
				i++
				names = append(names, fmt.Sprintf("%s%d", prefix, i))
			}
		}
	}
	return names
}

func (f *astTypeGenerator) generateTypeFromMapType(
	mapType *ast.MapType,
	targetPackagePath string,
	importMap map[string]string,
) MapType {
	keyDef := f.generateTypeFromExpr(mapType.Key, targetPackagePath, importMap)
	valDef := f.generateTypeFromExpr(mapType.Value, targetPackagePath, importMap)
	return MapType{
		Key:  keyDef,
		Elem: valDef,
	}
}

func (f *astTypeGenerator) generateTypeFromChanType(
	chanType *ast.ChanType,
	targetPackagePath string,
	importMap map[string]string,
) ChanType {
	c := f.generateTypeFromExpr(chanType.Value, targetPackagePath, importMap)

	switch chanType.Dir {
	case ast.RECV:
		return ChanType{Dir: ChanTypeDirRecv, Elem: c}
	case ast.SEND:
		return ChanType{Dir: ChanTypeDirSend, Elem: c}
	default:
		return ChanType{Dir: ChanTypeDirBoth, Elem: c}
	}
}

func (f *astTypeGenerator) generateTypeFromStructType(
	structType *ast.StructType,
	targetPackagePath string,
	importMap map[string]string,
) StructType {
	if structType.Fields == nil {
		return StructType{Fields: nil}
	}

	fields := make([]TypeField, 0, structType.Fields.NumFields())
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			fields = append(
				fields,
				TypeField{
					Name: name.String(),
					Type: f.generateTypeFromExpr(field.Type, targetPackagePath, importMap),
				},
			)
		}
	}

	return StructType{Fields: fields}
}

func (f *astTypeGenerator) generateTypeFromInterfaceType(
	interfaceType *ast.InterfaceType,
	targetPackagePath string,
	importMap map[string]string,
) InterfaceType {
	if interfaceType.Methods == nil {
		return InterfaceType{Methods: nil}
	}

	nMethod := interfaceType.Methods.NumFields()
	methods := make([]InterfaceTypeMethod, 0, nMethod)
	for _, field := range interfaceType.Methods.List {
		name := field.Names[0].String()
		funcType := field.Type.(*ast.FuncType)
		methods = append(methods, InterfaceTypeMethod{
			Name: name,
			Func: f.generateTypeFromFuncType(funcType, targetPackagePath, importMap),
		})
	}

	return InterfaceType{Methods: methods}
}
