Bugless architecture
====================

State: **draft** / current / obsolete

Requirements
------------

 - first class API support for end-users
 - pluggable model/storage backends
 - easily deployable in HA mode on k8s
 - ...


Diagram
-------

    .-------------.              .-------------------.
    | WebFrontend |       gRPC   | Model             |
    | (v0, Go)    |------------->| (switchable)      |
    |             |              '-------------------'
    |             |---.
    '-------------'    \  gRPC   .-------------------.
          .             '------->| Authenticator     |
          .                      | (switchable)      |
          .           ...        '-------------------'
    |             |  /
    | APIFrontend |-'
    | (Go)        |
    |             |
    '-------------'

Pretty simple. We currently split the design into three main components:

 - the **WebFrontend**, an HTTP service accessible by end users, serving the main web interface of the tracker.
 - the **APIFrontend**, a gateway for bots and scripts to perform actions on the tracker.
 - the **Model**, a nebulous service that provides the actual data storage (issues, components, hotlists)
 - the **Authenticator**, a nebulous service that allows people (either physical persons or bots) to authenticate and provides them with an identity. The Authenticator *might* also respond to authz/ACL requests.


Model considerations
--------------------

The model is not aware of users in any other aspect than an opaque identifier. This identifier should be easy to index on, and can refer to a person whose visible identity (email address, bot name or ...) can change.

The model keeps a flat, externally visible issue identifier. This identifier *might* be implemented as a [Snowflake identifier](https://developer.twitter.com/en/docs/basics/twitter-ids.html) on some model implementations.

WebFrontend considerations
--------------------------

The WebFrontend will have to join issues with user data from the authenticator. User data should either be cached in the frontend, or the authenticator should have its owns latency guarantees for accessing user information on every request (thus keeping a cache itself). I (q3k) think the second options is preferrable.

The WebFrontend should be stateless and horizontally scalable. Thus, special considerations must be made for HTTP request routing to always hit the appropriate WebFrontend if there are any incompatible HTML/JS changes being rolled out. Additionally, WebFrontends should have the ability to steer customers to another instance or version of the WebFrontend (for when an old deployment is being drained and turned down). We are targetting whatever the k8s Ingress model supports (preferably not being dependent on GCP or Nginx ingress extensions).


APIFrontend consideration
-------------------------

The APIFrontend is a service 'parallel' to the WebFrontend, providing similar functionality but via APIs. It *might* share some logic with the WebFrontend (and possibly other frontend), or might even be implemented in the same binary as the WebFrontend (and functionality can be selected by flags).

The APIFrontend will nearly directly expose Model APIs to the end-users. It should only respond to requests authenticated via TLS client certificates.

Model implementations
---------------------

Currently, we consider the following model implementations:

 - PostgreSQL-based
 - CockroachDB-based (**q3k** is currently researching this)

Authenticator implementations
-----------------------------

TODO :)

The gRPC Authenticator API should let at least three classes of flows be executed:

 - plaintext username/password query passed through the frontend to the authenticator
 - OAuth browser flow
 - TLS client certificate auth (with the Authenticator responding with user information based on a given TLS certificate CN)


Multi-region HA
---------------

An issue tracker is a core service of an organization. Thus, it needs to be runnable in a HA setup, preferably across regions with high latency links.

We consider the following high-level diagram:



       Region A         Region B         Region C    
    .-------------.  .-------------.  .-------------.
    |             |  |             |  |             |
     .----. .----.    .----. .----.    .----. .----. 
     | FE | | FE |    | FE | | FE |    | FE | | FE | 
     '----' '----'    '----' '----'    '----' '----' 
       ||     ||        ||     ||        ||     ||
       v|     v|        v|     v|        v|     v|
     .---------------------------------------------.
     |                    Model                    |
     '---------------------------------------------'
        |      |         |      |         |      |
        v      v         v      v         v      v
     .---------------------------------------------.
     |                Authenticator                |
     '---------------------------------------------'


Thus, we defer most of the cross-region logic to the Model and Authenticator :)

The rationale for this is that the two components are respectively best suited to implement their own HA and replication logic. For instance, the CockroachDB Model can generally leverage the database's multi-zone support, and an LDAP-based Authenticator will rely on an LDAP master and LDAP read replicas.
