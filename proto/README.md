# Athens external storage gRPC definitions

Protocol Buffers & gRPC Language References:

- [Language Guide](https://developers.google.com/protocol-buffers/docs/proto)
- [Full Language Specification](https://developers.google.com/protocol-buffers/docs/reference/proto3-spec)
- [gRPC homepage](https://grpc.io)

## How to generate code

Go client code gets generated from [`external_storage.proto`](./external_storage.proto) in this directory. You'll need the `protoc` generic code generator and the Go plugin for gRPC. See below for how to install those:

1. Install the `protoc` CLI.
   - On Mac, you can do this with [Homebrew](https://brew.sh) by running `brew install protobuf`
   - On other systems, you can download `protoc` directly by going to the by going to the [releases page](https://github.com/protocolbuffers/protobuf/releases) and downloading the right version for you. Don't download the language-specific ones or the `*-all*` one. Instead, just download the one specific to your platform.
2. Run `go install github.com/golang/protobuf/protoc-gen-go` to get the Protocol Buffers go plugin

After you've done that, you can re-generate the code with this command:

```bash
$ protoc -I $PWD external_storage.proto --go_out=$PWD/../pkg/storage/external
```

>Make sure to run this command from inside of this directory.
