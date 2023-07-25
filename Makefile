install:
	go build -o ~/go/bin/protoc-gen-connect-mock-server cmd/protoc-gen-connect-mock-server/main.go

generate: install
	buf generate