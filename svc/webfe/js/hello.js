// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

goog.provide("bugless");

goog.require("grpc.web.Status");
goog.require("proto.bugless.svc.ModelPromiseClient");

/**
 * The main Bugless application.
 *
 * @constructor
 * @param {!HTMLBodyElement} container Container to run app in.
 * @export
 */
bugless.App = function(container) {
    console.log("Starting Bugless...");
    /**
     * @type {!HTMLBodyElement}
     * @private
     * @const
     * @suppress {unusedPrivateMembers}
     */
    this.container_ = container;
    this.stub_ = new proto.bugless.svc.ModelPromiseClient("rpc", null, null);
};

/**
 * Foo.
 */
bugless.App.prototype.foo = function() {
    let req = new proto.bugless.svc.ModelGetIssuesRequest();
    req.setReturnType(proto.bugless.svc.ModelGetIssuesRequest.ReturnType.RETURNTYPE_FULL);
    let bySearch = new proto.bugless.svc.ModelGetIssuesRequest.BySearch();
    bySearch.setSearch("foo");
    req.setBySearch(bySearch);
    let stream = this.stub_.getIssues(req, null);
    stream.on('status', (/** @type {!grpc.web.Status.Status} */ status) => {
        console.log("status", status);
    });
};

/**
 * Main application entrypoint
 *
 * @param {!Element} container The container to bind the application to.
 * @export
 */
bugless.run = (container) => {
    let app = new bugless.App(/** @type {!HTMLBodyElement} */ (container));
    app.foo();
};
