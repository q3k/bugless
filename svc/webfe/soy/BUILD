load("@io_bazel_rules_closure//closure:defs.bzl", "closure_js_template_library")

closure_js_template_library(
    name = "soy",
    srcs = [
        "base.soy",
        "bugless.soy",
    ],
    visibility = ["//visibility:public"],
)

load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//extras:embed_data.bzl", "go_embed_data")

go_embed_data(
    name = "soy_data",
    srcs = [
        "base.soy",
        "bugless.soy",
    ],
    package = "soy",
    flatten = True,
)

# keep
go_library(
    name = "go_default_library",
    srcs = [":soy_data"],
    importpath = "github.com/q3k/bugless/svc/webfe/soy",
    visibility = ["//visibility:public"],
)
