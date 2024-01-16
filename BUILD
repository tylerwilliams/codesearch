load("@bazel_gazelle//:def.bzl", "gazelle")
load("//rules/go:index.bzl", "go_sdk_tool")

# gazelle:prefix github.com/google/codesearch
# gazelle:build_file_name BUILD,BUILD.bazel
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
