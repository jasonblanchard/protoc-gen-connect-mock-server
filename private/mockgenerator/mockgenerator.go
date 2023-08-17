package mockgenerator

import (
	"fmt"

	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Dependency struct {
	Messages       []*protogen.Message
	GoImportPath  string
	GoPackageName string
}

type GenerateFileInput struct {
	PkgName           string
	File              *protogen.File
	Filename          string
	GeneratedFile     *protogen.GeneratedFile
	Dependencies []*Dependency
}

func GenerateFile(i *GenerateFileInput) error {
	trimmedImportPath := i.File.GoImportPath.String()[1 : len(i.File.GoImportPath.String())-1]
	connectPkgName := string(i.File.GoPackageName) + "connect"
	pkgNamespace := string(i.File.GoPackageName)

	importsStringInput := &ImportsStringInput{
		PkgName:         i.PkgName,
		GoImportPath:    i.File.GoImportPath.String(),
		ConnectFilePath: trimmedImportPath + "/" + connectPkgName,
		Dependencies: []string{},
	}

	for _, dep := range i.Dependencies {
		importsStringInput.Dependencies = append(importsStringInput.Dependencies, dep.GoImportPath)
	}

	importsString, err := importsString(importsStringInput)

	if err != nil {
		return fmt.Errorf("error getting importsString: %w", err)
	}

	i.GeneratedFile.P(importsString)

	for _, dep := range i.Dependencies {
		for _, message := range dep.Messages {
			input := &MockMessageInput{
				Name:         string(message.Desc.Name()),
				PkgNamespace: dep.GoPackageName,
			}
	
			for _, field := range message.Fields {
				if field.Oneof == nil {
					input.Fields = append(input.Fields, Field{
						Name:  field.GoName,
						Value: fieldValueString(field, dep.GoPackageName, i.Dependencies),
					})
				}
				// } else {
				// 	// // TODO: This is adding one per oneOf variant
				// 	// input.Fields = append(input.Fields, Field{
				// 	// 	Name:  field.Oneof.GoName,
				// 	// 	Value: oneOfFieldValueString(field, dep.GoPackageName),
				// 	// })
				// }
			}
	
			messageString, err := mockMessageString(input)
	
			if err != nil {
				return fmt.Errorf("error getting mock message string: %w", err)
			}
	
			i.GeneratedFile.P(messageString)
			i.GeneratedFile.P("")
		}

	}

	for _, msg := range i.File.Messages {
		input := &MockMessageInput{
			Name:         string(msg.Desc.Name()),
			PkgNamespace: pkgNamespace,
		}

		for _, field := range msg.Fields {
			input.Fields = append(input.Fields, Field{
				Name:  field.GoName,
				Value: fieldValueString(field, pkgNamespace, i.Dependencies), // TODO: This isn't right, need to get which file it comes from
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
				PkgNamespace:   pkgNamespace,
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

func fieldValueString(field *protogen.Field, pkgNamespace string, deps []*Dependency) string {
	// TODO: Handle IsMap, oneOf

	// namspace
	// - Try to find it in deps by matching dep.GoImportPath == field.Message.GoIdent.GoImportPath.String()
	// - If match, use dep.GoImportPath
	// - If not, use pkgNamespace

	namespace := pkgNamespace

	matchedDependencyIndex := slices.IndexFunc(deps, func(d *Dependency) bool {
		if (field.Message == nil) {
			return false
		}
		return d.GoImportPath == field.Message.GoIdent.GoImportPath.String()
	})

	if (matchedDependencyIndex != -1) {
		namespace = deps[matchedDependencyIndex].GoPackageName
	}

	if field.Desc.IsList() {
		if field.Desc.Message() != nil {
			return "[]*" + pkgNamespace + "." + string(field.Message.Desc.Name()) + "{" + pkgNamespace + "_NewMock" + string(field.Message.Desc.Name()) + "()}"
		} else {
			return fmt.Sprintf("[]string{%v, %v, %v}", getStaticFieldValue(field, pkgNamespace), getStaticFieldValue(field, pkgNamespace), getStaticFieldValue(field, pkgNamespace))
		}
	} else {
		if field.Desc.Message() != nil {
			// return pkgNamespace + "_NewMock" + string(field.Message.Desc.Name()) + "()"
			return namespace + "_NewMock" + string(field.Message.Desc.Name()) + "()"
		} else {
			return getStaticFieldValue(field, pkgNamespace)
		}
	}
}

// func oneOfFieldValueString(field *protogen.Field, pkgNamespace string) string {
// 	return "&" + pkgNamespace + "." + field.Parent.GoIdent.GoName + "_" + field.GoName + "{}"
// }

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
