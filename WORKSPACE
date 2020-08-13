workspace(
    name = "bugless",
)

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

# Protobuf library

http_archive(
    name = "com_google_protobuf",
    sha256 = "9748c0d90e54ea09e5e75fb7fac16edce15d2028d4356f32211cfa3c0e956564",
    strip_prefix = "protobuf-3.11.4",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/v3.11.4.zip"],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

# Go & Gazelle rules

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "2697f6bc7c529ee5e6a2d9799870b9ec9eaeb3ee7d70ed50b87a2c2c97e13d9e",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.23.8/rules_go-v0.23.8.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.23.8/rules_go-v0.23.8.tar.gz",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "cdb02a887a7187ea4d5a27452311a75ed8637379a1287d8eeb952138ea485f7d",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

# Closure rules

http_archive(
    name = "io_bazel_rules_closure",
    sha256 = "d7a04263cf5b7af90f52d759da1e50c3cfe81c6cb16eec430af86e6bed248098",
    strip_prefix = "rules_closure-0e187366b658d1796d2580f8b7e1a8d7e7e1492d",
    urls = [
        "https://github.com/bazelbuild/rules_closure/archive/0e187366b658d1796d2580f8b7e1a8d7e7e1492d.zip",
    ],
)

load("@io_bazel_rules_closure//closure:repositories.bzl", "rules_closure_dependencies", "rules_closure_toolchains")

rules_closure_dependencies()

rules_closure_toolchains()

# gRPC-Web rules
http_archive(
    name = "com_github_grpc_grpc_web",
    sha256 = "283fdea0ff1539f47315700fc557837d539cbaa2b6e5e5dcb0b7280d3f054029",
    strip_prefix = "grpc-web-6b99a37519fd5de2d46f7fd4d1d293504e15161f",
    urls = [
        "https://github.com/grpc/grpc-web/archive/6b99a37519fd5de2d46f7fd4d1d293504e15161f.tar.gz",
    ],
)

# Go repositories

go_repository(
    name = "co_honnef_go_tools",
    commit = "c2f93a96b099",
    importpath = "honnef.co/go/tools",
)

go_repository(
    name = "com_github_burntsushi_toml",
    importpath = "github.com/BurntSushi/toml",
    tag = "v0.3.1",
)

go_repository(
    name = "com_github_client9_misspell",
    importpath = "github.com/client9/misspell",
    tag = "v0.3.4",
)

go_repository(
    name = "com_github_go_stack_stack",
    importpath = "github.com/go-stack/stack",
    tag = "v1.8.0",
)

go_repository(
    name = "com_github_golang_glog",
    commit = "23def4e6c14b",
    importpath = "github.com/golang/glog",
)

go_repository(
    name = "com_github_golang_mock",
    importpath = "github.com/golang/mock",
    tag = "v1.1.1",
)

go_repository(
    name = "com_github_golang_protobuf",
    importpath = "github.com/golang/protobuf",
    tag = "v1.3.1",
)

go_repository(
    name = "com_github_inconshreveable_log15",
    commit = "67afb5ed74ec",
    importpath = "github.com/inconshreveable/log15",
)

go_repository(
    name = "com_github_mattn_go_colorable",
    importpath = "github.com/mattn/go-colorable",
    tag = "v0.1.1",
)

go_repository(
    name = "com_github_mattn_go_isatty",
    importpath = "github.com/mattn/go-isatty",
    tag = "v0.0.7",
)

go_repository(
    name = "com_github_q3k_statusz",
    commit = "924f04ea7114",
    importpath = "github.com/q3k/statusz",
)

go_repository(
    name = "com_github_shirou_gopsutil",
    importpath = "github.com/shirou/gopsutil",
    tag = "v2.18.12",
)

go_repository(
    name = "com_google_cloud_go",
    importpath = "cloud.google.com/go",
    tag = "v0.26.0",
)

go_repository(
    name = "org_golang_google_appengine",
    importpath = "google.golang.org/appengine",
    tag = "v1.1.0",
)

go_repository(
    name = "org_golang_google_genproto",
    commit = "c66870c02cf8",
    importpath = "google.golang.org/genproto",
)

go_repository(
    name = "org_golang_google_grpc",
    importpath = "google.golang.org/grpc",
    tag = "v1.29.1",
)

go_repository(
    name = "org_golang_x_crypto",
    commit = "c2843e01d9a2",
    importpath = "golang.org/x/crypto",
)

go_repository(
    name = "org_golang_x_lint",
    commit = "d0100b6bd8b3",
    importpath = "golang.org/x/lint",
)


go_repository(
    name = "org_golang_x_net",
    importpath = "golang.org/x/net",
    sum = "h1:oWX7TPOiFAMXLq8o0ikBYfCJVlRHBcsciT5bXOrH628=",
    version = "v0.0.0-20190311183353-d8887717615a",
)

go_repository(
    name = "org_golang_x_oauth2",
    commit = "d2e6202438be",
    importpath = "golang.org/x/oauth2",
)

go_repository(
    name = "org_golang_x_sync",
    commit = "1d60e4601c6f",
    importpath = "golang.org/x/sync",
)

go_repository(
    name = "org_golang_x_sys",
    commit = "a9d3bda3a223",
    importpath = "golang.org/x/sys",
)

go_repository(
    name = "org_golang_x_text",
    importpath = "golang.org/x/text",
    sum = "h1:g61tztE5qeGQ89tm6NTjjM9VPIm088od1l6aSorWRWg=",
    version = "v0.3.0",
)

go_repository(
    name = "org_golang_x_tools",
    commit = "11955173bddd",
    importpath = "golang.org/x/tools",
)

go_repository(
    name = "com_github_jmoiron_sqlx",
    commit = "d7d95172beb5a538ff08c50a6e98aee6e7c41f40",
    importpath = "github.com/jmoiron/sqlx",
)

go_repository(
    name = "com_github_lib_pq",
    commit = "f91d3411e481ed313eeab65ebfe9076466c39d01",
    importpath = "github.com/lib/pq",
)

go_repository(
    name = "pl_hackerspace_code_hscloud",
    commit = "a01c487a6e16bdc7c2bef25e3414c3720368b7ca",
    importpath = "code.hackerspace.pl/hscloud",
)

go_repository(
    name = "com_github_golang_migrate_migrate_v4",
    commit = "8fd31348ae7b1941edb69e0b7ab372387051b4e2",
    importpath = "github.com/golang-migrate/migrate/v4",
    remote = "https://github.com/q3k/migrate",
    vcs = "git",
)

go_repository(
    name = "com_github_gchaincl_sqlhooks",
    commit = "1932c8dd22f2283687586008bf2d58c2c5c014d0",
    importpath = "github.com/gchaincl/sqlhooks",
)

go_repository(
    name = "io_k8s_client_go",
    commit = "14c42cd304d9e112a51d79937759edbd3665305b",
    importpath = "k8s.io/client-go",
)

go_repository(
    name = "com_github_hashicorp_go_multierror",
    commit = "bdca7bb83f603b80ef756bb953fe1dafa9cd00a2",
    importpath = "github.com/hashicorp/go-multierror",
)

go_repository(
    name = "com_github_hashicorp_errwrap",
    commit = "8a6fb523712970c966eefc6b39ed2c5e74880354",
    importpath = "github.com/hashicorp/errwrap",
)

go_repository(
    name = "io_k8s_apimachinery",
    commit = "6e68a40eebf94cb3e778d807bce5637d6aa079ea",
    importpath = "k8s.io/apimachinery",
    build_file_proto_mode = "disable",
)

go_repository(
    name = "io_k8s_utils",
    commit = "8d271d903fe4c290aa361acfb242cff7bcee96f1",
    importpath = "k8s.io/utils",
)

go_repository(
    name = "io_k8s_klog",
    commit = "3ca30a56d8a775276f9cdae009ba326fdc05af7f",
    importpath = "k8s.io/klog",
)

go_repository(
    name = "com_github_googleapis_gnostic",
    commit = "b0a17e38ce1aad0c792ef9efd1810364be151db4",
    importpath = "github.com/googleapis/gnostic",
)

go_repository(
    name = "io_k8s_api",
    commit = "816a9b7df6780e786246b6e46caadaed8d17e34c",
    importpath = "k8s.io/api",
    build_file_proto_mode = "disable",
)

go_repository(
    name = "org_golang_x_time",
    commit = "c4c64cad1fd0a1a8dab2523e04e61d35308e131e",
    importpath = "golang.org/x/time",
)

go_repository(
    name = "com_github_go_logr_logr",
    commit = "a1ebd699b1950beb1da0752cbb5559662018f798",
    importpath = "github.com/go-logr/logr",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    commit = "d8f796af33cc11cb798c1aaeb27a4ebc5099927d",
    importpath = "github.com/davecgh/go-spew",
)

go_repository(
    name = "com_github_google_gofuzz",
    commit = "b906efc57a556621b61db18d73df8c109dfa3613",
    importpath = "github.com/google/gofuzz",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    commit = "f221b8435cfb71e54062f6c6e99e9ade30b124d5",
    importpath = "gopkg.in/yaml.v2",
)

go_repository(
    name = "com_github_json_iterator_go",
    commit = "03217c3e97663914aec3faafde50d081f197a0a2",
    importpath = "github.com/json-iterator/go",
)

go_repository(
    name = "com_github_modern_go_reflect2",
    commit = "94122c33edd36123c84d5368cfb2b69df93a0ec8",
    importpath = "github.com/modern-go/reflect2",
)

go_repository(
    name = "io_k8s_sigs_yaml",
    commit = "4cd0c284b15f1735b8cc247df097d262b8903f9f",
    importpath = "sigs.k8s.io/yaml",
)

go_repository(
    name = "com_github_modern_go_concurrent",
    commit = "bacd9c7ef1dd9b15be4a9909b8ac7a4e313eec94",
    importpath = "github.com/modern-go/concurrent",
)

go_repository(
    name = "in_gopkg_inf_v0",
    commit = "d2d2541c53f18d2a059457998ce2876cc8e67cbf",
    importpath = "gopkg.in/inf.v0",
)

go_repository(
    name = "com_github_cockroachdb_cockroach_go_v2",
    importpath = "github.com/cockroachdb/cockroach-go/v2",
    sum = "h1:kG3i3YDEcg8Sby8PzUu3Wvp67Ienevj4bRAnHDF/xNc=",
    version = "v2.0.5",
)

go_repository(
    name = "com_github_jackc_pgconn",
    commit = "9449f4b08174c68107663a000f8eddc66263fca1",
    importpath = "github.com/jackc/pgconn",
)

go_repository(
    name = "com_github_jackc_pgpassfile",
    commit = "99d8e8e28945ffceaf75b0299fcb2bb656b8a683",
    importpath = "github.com/jackc/pgpassfile",
)

go_repository(
    name = "com_github_jackc_pgio",
    commit = "8d9c2a3dafd92d070bd758a165022fd1059e3195",
    importpath = "github.com/jackc/pgio",
)

go_repository(
    commit = "eca1e51822f3ebd8f48f62029d3e32f931d32c32",
    name = "com_github.com_jackc_pgproto3_v2",
    importpath = "github.com/jackc/pgproto3/v2",
    remote = "https://github.com/jackc/pgproto3",
    vcs = "git",
)

go_repository(
    name = "org_golang_x_xerrors",
    commit = "1b5146add8981d58be77b16229c0ff0f8bebd8c1",
    importpath = "golang.org/x/xerrors",
)

go_repository(
    name = "com_github_jackc_chunkreader_v2",
    commit = "2c463c0e7d0d0876517f087ce2cce66a46182141",
    importpath = "github.com/jackc/chunkreader/v2",
    remote = "https://github.com/jackc/chunkreader",
    vcs = "git",
)

go_repository(
    name = "com_github_robfig_soy",
    importpath = "github.com/robfig/soy",
    remote = "https://github.com/q3k/soy",
    commit = "c45262fef8da275270c3cf0cb781ed5d87188b38",
    vcs = "git",
)

go_repository(
    name = "com_github_improbable_eng_grpc_web",
    importpath = "github.com/improbable-eng/grpc-web",
    sum = "h1:GlCS+lMZzIkfouf7CNqY+qqpowdKuJLSLLcKVfM1oLc=",
    version = "v0.12.0",
)

go_repository(
    name = "com_github_gorilla_websocket",
    importpath = "github.com/gorilla/websocket",
    sum = "h1:+/TMaTYc4QFitKJxsQ7Yye35DkWvkdLcvGKqM+x0Ufc=",
    version = "v1.4.2",
)

go_repository(
    name = "com_github_rs_cors",
    importpath = "github.com/rs/cors",
    sum = "h1:+88SsELBHx5r+hZ8TCkggzSstaWNbDvThkVK8H6f9ik=",
    version = "v1.7.0",
)

go_repository(
    name = "com_github_desertbit_timer",
    importpath = "github.com/desertbit/timer",
    sum = "h1:U5y3Y5UE0w7amNe7Z5G/twsBW0KEalRQXZzf8ufSh9I=",
    version = "v0.0.0-20180107155436-c41aec40b27f",
)

go_repository(
    name = "com_github_blevesearch_bleve",
    importpath = "github.com/blevesearch/bleve",
    build_file_proto_mode = "disable",
    sum = "h1:4PspZE7XABMSKcVpzAKp0E05Yer1PIYmTWk+1ngNr/c=",
    version = "v1.0.7",
)

go_repository(
    name = "com_github_blevesearch_segment",
    importpath = "github.com/blevesearch/segment",
    sum = "h1:5lG7yBCx98or7gK2cHMKPukPZ/31Kag7nONpoBt22Ac=",
    version = "v0.9.0",
)

go_repository(
    name = "com_github_blevesearch_snowballstem",
    importpath = "github.com/blevesearch/snowballstem",
    sum = "h1:lMQ189YspGP6sXvZQ4WZ+MLawfV8wOmPoD/iWeNXm8s=",
    version = "v0.9.0",
)

go_repository(
    name = "com_github_blevesearch_go_porterstemmer",
    importpath = "github.com/blevesearch/go-porterstemmer",
    sum = "h1:GtmsqID0aZdCSNiY8SkuPJ12pD4jI+DdXTAn4YRcHCo=",
    version = "v1.0.3",
)

go_repository(
    name = "com_github_steveyen_gtreap",
    importpath = "github.com/steveyen/gtreap",
    sum = "h1:CjhzTa274PyJLJuMZwIzCO1PfC00oRa8d1Kc78bFXJM=",
    version = "v0.1.0",
)

go_repository(
    name = "io_etcd_go_bbolt",
    importpath = "go.etcd.io/bbolt",
    sum = "h1:hi1bXHMVrlQh6WwxAy+qZCV/SYIlqo+Ushwdpa4tAKg=",
    version = "v1.3.4",
)

go_repository(
    name = "com_github_blevesearch_zap_v12",
    importpath = "github.com/blevesearch/zap/v12",
    sum = "h1:y8FWSAYkdc4p1dn4YLxNNr1dxXlSUsakJh2Fc/r6cj4=",
    version = "v12.0.7",
)

go_repository(
    name = "com_github_couchbase_vellum",
    importpath = "github.com/couchbase/vellum",
    sum = "h1:qrj9ohvZedvc51S5KzPfJ6P6z0Vqzv7Lx7k3mVc2WOk=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_glycerine_go_unsnap_stream",
    importpath = "github.com/glycerine/go-unsnap-stream",
    sum = "h1:FQqoVvjbiUioBBFUL5up+h+GdCa/AnJsL/1bIs/veSI=",
    version = "v0.0.0-20190901134440-81cf024a9e0a",
)

go_repository(
    name = "com_github_blevesearch_zap_v11",
    importpath = "github.com/blevesearch/zap/v11",
    sum = "h1:nnmAOP6eXBkqEa1Srq1eqA5Wmn4w+BZjLdjynNxvd+M=",
    version = "v11.0.7",
)

go_repository(
    name = "com_github_tinylib_msgp",
    importpath = "github.com/tinylib/msgp",
    sum = "h1:gWmO7n0Ys2RBEb7GPYB9Ujq8Mk5p2U08lRnmMcGy6BQ=",
    version = "v1.1.2",
)

go_repository(
    name = "com_github_philhofer_fwd",
    importpath = "github.com/philhofer/fwd",
    sum = "h1:UbZqGr5Y38ApvM/V/jEljVxwocdweyH+vmYvRPBnbqQ=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_golang_snappy",
    importpath = "github.com/golang/snappy",
    sum = "h1:Qgr9rKW7uDUkrbSmQeiDsGa8SjGyCOGtuasMWwvp2P4=",
    version = "v0.0.1",
)

go_repository(
    name = "com_github_blevesearch_mmap_go",
    importpath = "github.com/blevesearch/mmap-go",
    sum = "h1:JtMHb+FgQCTTYIhtMvimw15dJwu1Y5lrZDMOFXVWPk0=",
    version = "v1.0.2",
)

go_repository(
    name = "com_github_willf_bitset",
    importpath = "github.com/willf/bitset",
    sum = "h1:NotGKqX0KwQ72NUzqrjZq5ipPNDQex9lo3WpaS8L2sc=",
    version = "v1.1.10",
)

go_repository(
    name = "com_github_roaringbitmap_roaring",
    importpath = "github.com/RoaringBitmap/roaring",
    sum = "h1:gpyfd12QohbqhFO4NVDUdoPOCXsyahYRQhINmlHxKeo=",
    version = "v0.4.23",
)

go_repository(
    name = "com_github_stackexchange_wmi",
    importpath = "github.com/stackexchange/wmi",
    sum = "h1:OKjqz+yB+g0IS7H6EuOYhs9XMAal3zEfHmV4GR8EGDc=",
    version = "v0.0.0-20190523213315-cbe66965904d",
)

go_repository(
    name = "com_github_go_ole_go_ole",
    importpath = "github.com/go-ole/go-ole",
    sum = "h1:nNBDSCOigTSiarFpYE9J/KtEA1IOW4CNeqT9TQDqCxI=",
    version = "v1.2.4",
)

go_repository(
    name = "com_github_pkg_errors",
    importpath = "github.com/pkg/errors",
    sum = "h1:FEBLx1zS214owpjy7qsBeixbURkuhQAwrK5UwLGTwt4=",
    version = "v0.9.1",
)
