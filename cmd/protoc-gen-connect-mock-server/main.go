package main

import (
	"github.com/jasonblanchard/protoc-gen-connect-mock-server/private/mockgenerator"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			filename := f.GeneratedFilenamePrefix + "connectmockserver/main.pb.go"
			g := gen.NewGeneratedFile(filename, f.GoImportPath)

			input := &mockgenerator.GenerateFileInput{
				PkgName: "protoc-gen-connect-mock-server",
				File: f,
				Filename: filename,
				GeneratedFile: g,
			}
			err := mockgenerator.GenerateFile(input)

			if err != nil {
				panic(err) // TODO: Better error handling?
			}
		}
		return nil
	})
}