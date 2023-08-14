package module

import (
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
)

func TestExampleCreateFileString(t *testing.T) {
	file := &protogen.File{
		GoImportPath: "github.com/jasonblanchard/protoc-gen-connect-mock-server/gen/greet/v1",
		GoPackageName: "greetv1",
		Messages: []*protogen.Message{
			{},
		},
	}

	result, err := CreateFileString(file)

	if err != nil {
		t.Errorf("error: %d", err)
	}

	t.Log(result)
}