package mockgenerator

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type GenerateFileInput struct {
	PkgName       string
	File          *protogen.File
	Filename      string
	GeneratedFile *protogen.GeneratedFile
}

func GenerateFile(i *GenerateFileInput) error {
	trimmedImportPath := i.File.GoImportPath.String()[1 : len(i.File.GoImportPath.String())-1]
	connectPkgName := string(i.File.GoPackageName) + "connect"
	pkgPathSlice := strings.Split(trimmedImportPath, "/")
	pkgNamespace := pkgPathSlice[len(pkgPathSlice)-1]

	importsStringInput := &ImportsStringInput{
		PkgName:         i.PkgName,
		GoImportPath:    trimmedImportPath,
		ConnectFilePath: trimmedImportPath + "/" + connectPkgName,
		PkgNamespace:    pkgNamespace,
	}

	importsString, err := importsString(importsStringInput)

	if err != nil {
		return fmt.Errorf("error getting importsString: %w", err)
	}

	i.GeneratedFile.P(importsString)

	for _, msg := range i.File.Messages {
		input := &MockMessageInput{
			Name: string(msg.Desc.Name()),
		}

		for _, field := range msg.Fields {
			input.Fields = append(input.Fields, Field{
				Name:  field.GoName,
				Value: fieldValueString(field, pkgNamespace),
			})
		}

		messageString, err := mockMessageString(input)

		if err != nil {
			return fmt.Errorf("error getting mock message string: %w", err)
		}

		i.GeneratedFile.P(messageString)
		i.GeneratedFile.P("")
	}

	for _, service := range i.File.Services {
		input := &MockServerInput{
			GoName: service.GoName,
		}

		for _, method := range service.Methods {
			input.Methods = append(input.Methods, MockMethod{
				MockServerName: service.GoName + "MockServer",
				GoName:         method.GoName,
				InputName:      string(method.Desc.Input().Name()),
				OutputName:     string(method.Desc.Output().Name()),
			})
		}

		serverString, err := mockServerString(input)

		if err != nil {
			return fmt.Errorf("error creating mock server string: %w", err)
		}

		i.GeneratedFile.P(serverString)
	}

	mainFuncStringInput := &MainFuncStringInput{
		Handlers: []MainFuncHandler{},
	}

	for _, service := range i.File.Services {
		mainFuncStringInput.Handlers = append(mainFuncStringInput.Handlers, MainFuncHandler{
			ConnectPkg:  connectPkgName,
			ServiceName: service.GoName,
		})
	}

	mainFuncString, err := mainFuncString(mainFuncStringInput)

	if err != nil {
		return fmt.Errorf("error creating main func: %w", err)
	}

	i.GeneratedFile.P("")
	i.GeneratedFile.P(mainFuncString)

	return nil
}

func fieldValueString(field *protogen.Field, pkgNamespace string) string {
	// TODO: Handle IsMap

	if field.Desc.IsList() {
		if field.Desc.Message() != nil {
			return "[]*v1." + string(field.Message.Desc.Name()) + "{NewMock" + string(field.Message.Desc.Name()) + "()}"
		} else {
			return fmt.Sprintf("[]string{%v, %v, %v}", getStaticFieldValue(field, pkgNamespace), getStaticFieldValue(field, pkgNamespace), getStaticFieldValue(field, pkgNamespace))
		}
	} else {
		if field.Desc.Message() != nil {
			return "NewMock" + field.GoName + "()"
		} else {
			return getStaticFieldValue(field, pkgNamespace)
		}
	}
}

func getStaticFieldValue(field *protogen.Field, pkgNamespace string) string {
	// TODO: GroupKind?
	switch field.Desc.Kind() {
	case protoreflect.EnumKind:
		return fmt.Sprintf("%s.%s_%s", pkgNamespace, field.GoName, field.Enum.Values[len(field.Enum.Values)-1].Desc.Name())
	case protoreflect.BoolKind:
		return "false"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind, protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind, protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind, protoreflect.FloatKind, protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind, protoreflect.DoubleKind:
		return "123"
	case protoreflect.StringKind:
		return "\"string\""
	case protoreflect.BytesKind:
		return "[]byte{1,2,3}"
	default:
		return "\"UNIMPLEMENTED\""
	}
}
