load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "csearch_lib",
    srcs = ["csearch.go"],
    importpath = "github.com/google/codesearch/cmd/csearch",
    visibility = ["//visibility:private"],
    deps = [
        "//index2",
        "//query",
        "//regexp",
    ],
)

go_binary(
    name = "csearch",
    embed = [":csearch_lib"],
    visibility = ["//visibility:public"],
)
