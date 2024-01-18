load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "index2",
    srcs = [
        "common.go",
        "mmap_bsd.go",
        "mmap_linux.go",
        "mmap_windows.go",
        "read.go",
        "write.go",
    ],
    importpath = "github.com/google/codesearch/index2",
    visibility = ["//visibility:public"],
    deps = [
        "//query",
        "//sparse",
        "@com_github_cockroachdb_pebble//:pebble",
        "@com_github_roaringbitmap_roaring//:roaring",
        "@org_golang_x_sync//errgroup",
    ],
)

go_test(
    name = "index2_test",
    srcs = ["index2_test.go"],
    deps = [":index2"],
)
