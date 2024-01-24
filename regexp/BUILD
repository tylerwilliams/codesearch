load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "regexp",
    srcs = [
        "copy.go",
        "match.go",
        "regexp.go",
        "utf.go",
    ],
    importpath = "github.com/google/codesearch/regexp",
    visibility = ["//visibility:public"],
    deps = [
        "//result",
        "//sparse",
    ],
)

go_test(
    name = "regexp_test",
    srcs = ["regexp_test.go"],
    embed = [":regexp"],
)
