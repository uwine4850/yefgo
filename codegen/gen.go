package codegen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

type ModuleTemplate struct {
	GoPkg    string
	FileName string
	Import   []string
	Classes  []ClassTemplate
}

type ArgsClassTemplate struct {
	StructName  string
	Args        string
	ArgsForFunc string
}

type ClassTemplate struct {
	Template  *template.Template
	Args      ArgsClassTemplate
	Functions []FuncTemplate
}

type ArgsFuncTemplate struct {
	StructName string
	GoFuncName string
	FuncCall   string
	Args       string
	OutputType string
}

type FuncTemplate struct {
	Template *template.Template
	Args     ArgsFuncTemplate
}

func yamlArgsToString(args []Arg) string {
	builder := strings.Builder{}
	for i := 0; i < len(args); i++ {
		if i == len(args)-1 {
			builder.WriteString(args[i].Name + " " + args[i].Type)
		} else {
			builder.WriteString(args[i].Name + " " + args[i].Type + ", ")
		}
	}

	return builder.String()
}

func yamlArgsAsFuncArgs(args []Arg) string {
	builder := strings.Builder{}
	for i := 0; i < len(args); i++ {
		if i == len(args)-1 {
			builder.WriteString(args[i].Name)
		} else {
			builder.WriteString(args[i].Name + ", ")
		}
	}
	return builder.String()
}

func yamlImportToString(goModule string, imports []string) string {
	builder := strings.Builder{}
	for i := 0; i < len(imports); i++ {
		builder.WriteString(fmt.Sprintf(". \"%s/%s\"", goModule, imports[i]))
	}
	return builder.String()
}

func capitalizeFuncName(name string) string {
	if len(name) == 0 {
		return name
	}
	firstChar := rune(name[0])
	if unicode.IsLetter(firstChar) {
		firstChar = unicode.ToUpper(firstChar)
	}
	return string(firstChar) + name[1:]
}

func createFuncTemplate(funcName string, className string, funcData Method) FuncTemplate {
	var funcCall bytes.Buffer
	if funcData.Type == "class" {
		funcCallTemp := template.Must(template.New("goFile").Parse(funcClassCall))
		err := funcCallTemp.Execute(&funcCall, map[string]string{"PyFuncName": funcName, "ArgsForFunc": yamlArgsAsFuncArgs(funcData.Args)})
		if err != nil {
			panic(err)
		}
	}
	if funcData.Type == "instance" {
		funcCallTemp := template.Must(template.New("goFile").Parse(funcInstanceCall))
		err := funcCallTemp.Execute(&funcCall, map[string]string{"PyFuncName": funcName, "ArgsForFunc": yamlArgsAsFuncArgs(funcData.Args)})
		if err != nil {
			panic(err)
		}
	}
	funcArgs := ArgsFuncTemplate{
		StructName: className,
		GoFuncName: capitalizeFuncName(funcName),
		Args:       yamlArgsToString(funcData.Args),
		OutputType: funcData.Output,
		FuncCall:   funcCall.String(),
	}
	var funcTemp *template.Template
	if funcData.Output == "" {
		funcTemp = template.Must(template.New("goFile").Parse(funcInit))
	} else {
		funcTemp = template.Must(template.New("goFile").Parse(funcWithOutputInit))
	}
	funcTemplate := FuncTemplate{
		funcTemp,
		funcArgs,
	}
	return funcTemplate
}

func executeTemplates(moduleTemplates []ModuleTemplate, goModuleName string) error {
	for j := 0; j < len(moduleTemplates); j++ {
		dirPath := filepath.Join("gen", moduleTemplates[j].GoPkg)
		if exist, err := pathExist(dirPath); err == nil && !exist {
			err := os.MkdirAll(dirPath, 0777)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		file, err := os.Create(filepath.Join(dirPath, moduleTemplates[j].FileName))
		if err != nil {
			return err
		}

		initPkg := template.Must(template.New("init").Parse(pkgInit))
		err = initPkg.Execute(file, map[string]string{"PkgName": moduleTemplates[j].GoPkg,
			"Imports": yamlImportToString(goModuleName+"/gen", moduleTemplates[j].Import)})
		if err != nil {
			return err
		}

		_classTemplates := moduleTemplates[j].Classes
		for i := 0; i < len(_classTemplates); i++ {
			err = _classTemplates[i].Template.Execute(file, _classTemplates[i].Args)
			if err != nil {
				return err
			}
			for j := 0; j < len(_classTemplates[i].Functions); j++ {
				err = _classTemplates[i].Functions[j].Template.Execute(file, _classTemplates[i].Functions[j].Args)
				if err != nil {
					return err
				}
			}
		}
		err = file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func Generate(cfgPath string, goModuleName string) error {
	cfg, err := ParseYefCfg(cfgPath)
	if err != nil {
		return err
	}
	var moduleTemplates []ModuleTemplate
	for _, class := range cfg.Modules {
		var classTemplates []ClassTemplate
		for i := 0; i < len(class.Classes); i++ {
			var args = ArgsClassTemplate{
				StructName:  class.Classes[i].Name,
				Args:        yamlArgsToString(class.Classes[i].Args),
				ArgsForFunc: yamlArgsAsFuncArgs(class.Classes[i].Args),
			}
			var classFunctions []FuncTemplate
			for funcName, funcData := range class.Classes[i].Methods {
				funcTemplate := createFuncTemplate(funcName, class.Classes[i].Name, funcData)
				classFunctions = append(classFunctions, funcTemplate)
			}
			classTemplates = append(classTemplates, ClassTemplate{
				Template:  template.Must(template.New("goFile").Parse(classinit)),
				Args:      args,
				Functions: classFunctions,
			})
		}
		moduleTemplate := ModuleTemplate{
			GoPkg:    class.GoPkg,
			FileName: class.FileName,
			Import:   class.Import,
			Classes:  classTemplates,
		}
		moduleTemplates = append(moduleTemplates, moduleTemplate)
	}
	err = executeTemplates(moduleTemplates, goModuleName)
	if err != nil {
		return err
	}
	return nil
}

func pathExist(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}
