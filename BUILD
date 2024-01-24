load("@bazel_gazelle//:def.bzl", "gazelle")
load("//rules/go:index.bzl", "go_sdk_tool")

# gazelle:prefix github.com/google/codesearch
# gazelle:build_file_name BUILD,BUILD.bazel
# gazelle:exclude testrepo/*
# gazelle:exclude **/*.pb.go
# gazelle:exclude bundle.go
# gazelle:prefix github.com/google/codesearch
# gazelle:proto disable
#
# Make these the default compilers for proto rules.
# See https://github.com/bazelbuild/rules_go/pull/3761 for more details
# gazelle:go_grpc_compilers     @io_bazel_rules_go//proto:go_proto,@io_bazel_rules_go//proto:go_grpc_v2"
# gazelle:go_proto_compilers	@io_bazel_rules_go//proto:go_proto,@io_bazel_rules_go//proto:go_grpc_v2
gazelle(
    name = "gazelle",
)

go_sdk_tool(
    name = "go",
    goroot_relative_path = "bin/go",
)

# Example usage: "bazel run //:gofmt -- -w ."
go_sdk_tool(
    name = "gofmt",
    goroot_relative_path = "bin/gofmt",
)
