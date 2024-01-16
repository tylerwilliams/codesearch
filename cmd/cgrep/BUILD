load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "cgrep_lib",
    srcs = ["cgrep.go"],
    importpath = "github.com/google/codesearch/cmd/cgrep",
    visibility = ["//visibility:private"],
    deps = ["//regexp"],
)

go_binary(
    name = "cgrep",
    embed = [":cgrep_lib"],
    visibility = ["//visibility:public"],
)
