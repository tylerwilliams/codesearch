load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "index",
    srcs = [
        "merge.go",
        "mmap_bsd.go",
        "mmap_linux.go",
        "mmap_windows.go",
        "read.go",
        "write.go",
    ],
    importpath = "github.com/google/codesearch/index",
    visibility = ["//visibility:public"],
    deps = [
        "//query",
        "//sparse",
    ],
)

go_test(
    name = "index_test",
    srcs = [
        "merge_test.go",
        "read_test.go",
        "write_test.go",
    ],
    embed = [":index"],
)
