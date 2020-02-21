# Athens external storage gRPC definitions

Generating code:

1. Install the `protoc` CLI.
   - On Mac, you can do this with [Homebrew](https://brew.sh) by running `brew install protobuf`
   - On other systems, you can download `protoc` directly by going to the by going to the [releases page](https://github.com/protocolbuffers/protobuf/releases) and downloading the right version for you. Don't download the language-specific ones or the `*-all*` one. Instead, just download the one specific to your platform.
2. Run `go install github.com/golang/protobuf/protoc-gen-go` to get the Protocol Buffers go plugin
3. Run `protoc -I $PWD external_storage.proto --go_out=$PWD/../pkg/storage/external` from inside this directory to generate the code
