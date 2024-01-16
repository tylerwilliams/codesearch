load("@bazel_gazelle//:deps.bzl", "go_repository")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

# bazelisk run //:gazelle -- update-repos -from_file=go.mod -to_macro=deps.bzl%install_go_mod_dependencies
def install_go_mod_dependencies(workspace_name = "codesearch"):
    pass
