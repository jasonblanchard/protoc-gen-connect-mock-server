package main

import (
	"github.com/jasonblanchard/protoc-gen-connect-mock-server/private/mockgenerator"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		dependencies := []*mockgenerator.Dependency{}

		for _, f := range gen.Files {
			if !f.Generate {

				dep := &mockgenerator.Dependency{
					GoImportPath:  f.GoImportPath.String(),
					GoPackageName: string(f.GoPackageName),
					Messages:      append([]*protogen.Message{}, f.Messages...),
				}

				dependencies = append(dependencies, dep)
				continue
			}

			filename := f.GeneratedFilenamePrefix + "connectmockserver/main.pb.go"
			g := gen.NewGeneratedFile(filename, f.GoImportPath)

			input := &mockgenerator.GenerateFileInput{
				PkgName:       "protoc-gen-connect-mock-server",
				File:          f,
				Filename:      filename,
				GeneratedFile: g,
				Dependencies:  dependencies,
			}
			err := mockgenerator.GenerateFile(input)

			if err != nil {
				panic(err) // TODO: Better error handling?
			}
		}
		return nil
	})
}
