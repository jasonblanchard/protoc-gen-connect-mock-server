package mockgenerator

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

type GenerateFileInput struct {
	PkgName string
	File *protogen.File
	Filename string
	G *protogen.GeneratedFile
}

func GenerateFile(i *GenerateFileInput) error {
	trimmedImportPath := i.File.GoImportPath.String()[1:len(i.File.GoImportPath.String())-1]

	importsStringInput := &ImportsStringInput{
		PkgName: i.PkgName,
		GoImportPath: trimmedImportPath,
		ConnectFilePath: trimmedImportPath + "/" + string(i.File.GoPackageName) + "connect",
	}

	importsString, err := importsString(importsStringInput)

	if err != nil {
		return fmt.Errorf("error getting importsString: %w", err)
	}
	
	i.G.P(importsString)

	for _, msg := range i.File.Messages {
		i.G.P("")
		i.G.P("func NewMock", msg.Desc.Name(), "() *v1.", msg.Desc.Name(), " {")
		i.G.P("mock := &v1.", msg.Desc.Name(), "{")
		for _, field := range msg.Fields {
			// TODO: Handle IsMap, Enum()

			if field.Desc.IsList() {
				if field.Desc.Message() != nil {
					i.G.P(field.GoName, ": ", "[]*v1.", field.Message.Desc.Name(), "{NewMock", field.Message.Desc.Name(), "()},")
				} else {
					// Otherwise, use mock scalar value
					i.G.P(field.GoName, ": ", "[]string{\"chello\", \"chello\", \"chello\"},") // TODO: Handle more than strings
				}
			} else {
				if field.Desc.Message() != nil {
					i.G.P(field.GoName, ": NewMock", field.GoName, "(),")
				} else {
					// Otherwise, use mock scalar value
					i.G.P(field.GoName, ": ", "\"chello\",") // TODO: Handle more than strings
				}
			}
		}
		i.G.P("}")
		i.G.P("return mock")
		i.G.P("}")
		i.G.P("")
	}

	for _, service := range i.File.Services {
		i.G.P("")
		i.G.P("type ", service.GoName, "MockServer ", "struct{}")
		i.G.P("")
		i.G.P("")
		for _, method := range service.Methods {
			i.G.P("func (", service.GoName, "MockServer) ", method.GoName, "(context.Context, *connect_go.Request[v1.", method.Desc.Input().Name(), "]) (*connect_go.Response[v1.", method.Desc.Output().Name(), "], error) {")
			i.G.P("resp := &connect_go.Response[v1.", method.Desc.Output().Name(), "]", "{}")
			i.G.P("resp.Msg = NewMock", method.Desc.Output().Name(), "()")
			i.G.P("return resp, nil")
			i.G.P("}")
		}
	}

	i.G.P("")
	i.G.P("func main() {")
	i.G.P("mux := http.NewServeMux()")
	for _, service := range i.File.Services {
		i.G.P("server := &", service.GoName, "MockServer{}")
		i.G.P("path, handler := greetv1connect.NewGreetServiceHandler(server)")
		i.G.P("mux.Handle(path, handler)")
	}
	
	i.G.P("corsHandler := cors.New(cors.Options{")
	i.G.P("AllowedOrigins: []string{\"https://buf.build\"},")
	i.G.P("AllowCredentials: true,")
	i.G.P("AllowedMethods:   []string{http.MethodPost, http.MethodOptions},")
	i.G.P("AllowedHeaders: []string{\"*\"},")
	i.G.P("}).Handler(mux)")
	i.G.P("http.ListenAndServe(\":8080\", h2c.NewHandler(corsHandler, &http2.Server{}))")
	i.G.P("}")

	return nil
}
