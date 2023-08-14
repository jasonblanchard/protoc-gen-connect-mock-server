package module

import (
	"bytes"
	"fmt"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

type TemplateModel struct {
	PkgName string
	File *protogen.File
	ConnectFilePath string
}

func registerTemplates() (*template.Template, error) {
	tmpls := map[string]string{
		"method": methodTpl,
		"service": serviceTpl,
		"file": fileTpl,
	}

	template := template.New("root")
	var err error

	for name, tmpl := range(tmpls) {
		template, err = template.New(name).Parse(tmpl)
		if err != nil {
			return template, fmt.Errorf("error parsing template: %w", err)
		}
	}

	return template, nil
}

func CreateFileString(file *protogen.File) (string, error) {
	tmpl, err := registerTemplates()

	if err != nil {
		return "", fmt.Errorf("error regiestering templates: %w", err)
	}

	buf := bytes.NewBufferString("")

	model := &TemplateModel{
		PkgName: "protoc-gen-connect-mock-server",
		File: file,
		ConnectFilePath: file.GoImportPath.String()[1:len(file.GoImportPath.String())-1] + "/" + string(file.GoPackageName) + "connect",
	}

	tmpl.Execute(buf, model)

	result := buf.String()
	
	return result, nil
}
