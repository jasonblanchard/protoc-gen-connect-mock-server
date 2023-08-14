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

			filename := f.GeneratedFilenamePrefix + "_connect_mock_server/main.pb.go"
			g := gen.NewGeneratedFile(filename, f.GoImportPath)

			// fileString, err := module.CreateFileString(f)

			// if (err != nil) {
			// 	return fmt.Errorf("error creating file string: %w", err)
			// }

			// g.P(fileString)

			// generateFile(gen, f, filename, g)
			input := &mockgenerator.GenerateFileInput{
				PkgName: "protoc-gen-connect-mock-server",
				File: f,
				Filename: filename,
				G: g,
			}
			err := mockgenerator.GenerateFile(input)

			if err != nil {
				panic(err) // TODO: Better error handling?
			}
		}
		return nil
	})
}