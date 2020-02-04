load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//extras:embed_data.bzl", "go_embed_data")
load("@io_bazel_rules_closure//closure:defs.bzl", "closure_css_library", "closure_css_binary")

closure_css_library(
    name = "gss_lib",
    srcs = [
        "bugless.gss",
    ],
)

closure_css_binary(
    name = "gss",
    deps = [
        ":gss_lib",
    ],
    visibility = ["//visibility:public"],
    renaming = False,
)

go_embed_data(
    name = "gss_data",
    srcs = [":gss"],
    package = "gss",
    flatten = True,
)

# keep
go_library(
    name = "go_default_library",
    srcs = [":gss_data"],
    importpath = "github.com/q3k/bugless/svc/webfe/gss",
    visibility = ["//visibility:public"],
)
