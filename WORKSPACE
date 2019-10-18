workspace(
    name = "bugless",
)

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "com_google_protobuf",
    sha256 = "33cba8b89be6c81b1461f1c438424f7a1aa4e31998dbe9ed6f8319583daac8c7",
    strip_prefix = "protobuf-3.10.0",
    urls = ["https://github.com/google/protobuf/archive/v3.10.0.zip"],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

http_archive(
    name = "io_bazel_rules_go",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/rules_go/releases/download/v0.20.1/rules_go-v0.20.1.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.20.1/rules_go-v0.20.1.tar.gz",
    ],
    sha256 = "842ec0e6b4fbfdd3de6150b61af92901eeb73681fd4d185746644c338f51d4c0",
)

http_archive(
    name = "bazel_gazelle",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/bazel-gazelle/releases/download/v0.19.0/bazel-gazelle-v0.19.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.19.0/bazel-gazelle-v0.19.0.tar.gz",
    ],
    sha256 = "41bff2a0b32b02f20c227d234aa25ef3783998e5453f7eade929704dcff7cd4b",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

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
    tag = "v1.20.1",
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
    commit = "d8887717615a",
    importpath = "golang.org/x/net",
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
    tag = "v0.3.0",
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
    commit = "e93eaeb3fe21ce2ccc1365277a01863e6bc84d9c",
    importpath = "github.com/golang-migrate/migrate/v4",
    remote = "https://github.com/golang-migrate/migrate",
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
    name = "com_github_cockroachdb_cockroach_go",
    commit = "606b3d062051259eca584a3f998d37e39b3b7622",
    importpath = "github.com/cockroachdb/cockroach-go",
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
