load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//extras:embed_data.bzl", "go_embed_data")
load("@io_bazel_rules_closure//closure:defs.bzl", "closure_js_binary", "closure_js_library")

closure_js_library(
    name = "js_lib",
    srcs = [
        "hello.js",
    ],
    deps = [
        "@io_bazel_rules_closure//closure/library",
        "@io_bazel_rules_closure//closure/library/ui:zippy",
        "//svc/webfe/soy",
    ],
)

closure_js_binary(
    name = "js",
    deps = [
        ":js_lib",
    ],
    entry_points = [
        "goog:bugless",
    ],
)

go_embed_data(
    name = "js_data",
    srcs = [":js"],
    package = "js",
    flatten = True,
)

# keep
go_library(
    name = "go_default_library",
    srcs = [":js_data"],
    importpath = "github.com/q3k/bugless/svc/webfe/js",
    visibility = ["//visibility:public"],
)
