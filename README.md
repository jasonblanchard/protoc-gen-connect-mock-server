# protoc-gen-connect-mock-server

Generate a functional mock server from a protobuf specification.

## Usage

Make sure the `protoc-gen-connect-mock-server` binary is compiled and in your `$PATH`.

Add it as a plugin to your `buf.gen.yaml` file along with `buf.build/protocolbuffers/go` and `buf.build/bufbuild/connect-go`:

```yaml
version: v1
managed:
  enabled: true
  go_package_prefix:
    default: <your Go module>
plugins:
  - plugin: buf.build/protocolbuffers/go
    out: gen
    opt: paths=source_relative
  - plugin: buf.build/bufbuild/connect-go
    out: gen
    opt: paths=source_relative
  - plugin: connect-mock-server
    out: gen
    opt: paths=source_relative
```

Create a protobuf service at `eliza/v1/eliza.proto` like this:

```
syntax = "proto3";

package eliza.v1;

message SayRequest {
  string sentence = 1;
}

message SayResponse {
  string sentence = 1;
}

service ElizaService {
  rpc Say(SayRequest) returns (SayResponse) {}
}
```

Generate the build:

```bash
$ buf build
```

Run the mock server

```bash
$ go run gen/eliza/v1/elizaconnectmockserver/main.pb.go
```

Test the service:

```bash
$ buf curl --schema eliza/v1/eliza.proto --data '{"sentence": "hello"}' http://localhost:8080/eliza.v1.ElizaService/Say
```

You should get a response like this:

```
{
  "sentence": "string"
}
```

Run it in dynamic mode with the `-d` flag

```bash
$ go run gen/eliza/v1/elizaconnectmockserver/main.pb.go -d
```

Running the same `buf curl` command above should return a different result each time like this:

```
{
  "sentence": "QDcyVRjMUAhhPMpSPwSXxxgMx"
}
```

## Status: Very Unstable

Pretty much everything in flight. The code is a mess. APIs can change anytime.
