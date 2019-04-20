Bugless architecture
====================

State: **draft** / current / obsolete

Requirements
------------

 - first class API support for end-users
 - pluggable model/storage backends
 - ...


Diagram
-------

(spongedown-compatible)

    .-------------.              .-------------------.
    | WebFrontend |       gRPC   | Model             |
    | (v0, Go)    |------------->| (switchable)      |
    |             |              '-------------------'
    |             |---.
    '-------------'    \  gRPC   .-------------------.
                        '------->| Authenticator     |
                                 | (switchable)      |
                                 '-------------------'

Pretty simple. We currently split the design into three main components:

 - the **WebFrontend**, an HTTP service accessible by end users, serving the main web interface of the tracker.
 - the **Model**, a nebulous service that provides the actual data storage (issues, components, hotlists)
 - the **Authenticator**, a nebulous service that allows people (either physical persons or bots) to authenticate and provides them with an identity. The Authenticator *might* also respond to authz/ACL requests.


Model considerations
--------------------

The model is not aware of users in any other aspect than an opaque identifier. This identifier should be easy to index on, and can refer to a person whose visible identity (email address, bot name or ...) can change.

The model keeps a flat, externally visible issue identifier. This identifier *might* be implemented as a [Snowflake identifier](https://developer.twitter.com/en/docs/basics/twitter-ids.html) on some model implementations.

WebFrontend considerations
--------------------------

The WebFrontend will have to join issues with user data from the authenticator. User data should either be cached in the frontend, or the authenticator should have its owns latency guarantees for accessing user information on every request (thus keeping a cache itself). I (q3k) think the second options is preferrable.

Model implementations
---------------------

Currently, we consider the following model implementations:

 - PostgreSQL-based
 - CockroachDB-based (**q3k** is currently researching this)

Authenticator implementations
-----------------------------

TODO :)

The gRPC Authenticator API should let at least two classes of flows be executed:

 - plaintext username/password query passed through the frontend to the authenticator
 - OAuth browser flow

