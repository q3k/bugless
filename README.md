bugless
=======

An experimental no-nonsense issue/bug tracker.

![screenshot](https://q3k.org/u/2ab3c1919cbac35c9a81f0788cf7d7e51a1cbeb1f2cf271b04b7ce9955db4100.png)

Status
------

In development. Ready some day, maybe.

Currently we have a CockroachDB backend/model, a frontend, and are able to list
issues.

**Stability and forward compatibilty**: this is pre-alpha software. All APIs,
schemas, and assumptions are subject to change - **even ones that are currently
written down as guarantees**. Once the software starts hitting alpha releases,
things will slowly begin to stabilize.


Building and running development instance
-----------------------------------------

Start the model with an in-memory database:

    bazel build //svc/model/crdb
    bazel-bin/svc/model/crdb/*/crdb -hspki_disable -eat_my_data

The model will listen on `:4200` for gRPC and `:4201` for debug HTTP.

Start the web frontend (this currently has a hard dep on an OIDC provider - any OIDC provider should do, but bugless is being actively developed against [sso.hackerspace.pl](https://sso.hackerspace.pl/)).:

    bazel build //svc/webfe
    bazel-bin/svc/webfe/*/webfe -hspki_disable -oidc_provider https://XXX -oidc_client_id YYY -oidc_client_secret ZZZ -secret hackme

The web frontend will connect to the model at `127.0.0.1:4200` by default,
serve debug HTTP at `:4211`, and serve public HTTP at `:8080`. Visting
`127.0.0.1:8080` with your browser should show you the bugless UI.

To add ann issue via [grpcurl](https://github.com/fullstorydev/grpcurl):

    grpcurl -plaintext -format=text -d \
        'author < id: "q3k@q3k.org" > initial_state < priority: 2 type: 1 status: 1 title: "Hello, World" > initial_comment: "Testing"' \
        127.0.0.1:4200 bugless.svc.Model.NewIssue


License
-------

Unless noted otherwise, this repository's contents are licensed under the GNU
Affero General Public LIcense v3.0 (or later), see [LICENSE](LICENSE).

