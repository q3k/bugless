// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

goog.provide("bugless");
goog.provide("bugless.notepad");

goog.require("bugless.templates.bugless");

goog.require("goog.dom");
goog.require("goog.events");
goog.require("goog.soy");


/**
 * A note.
 *
 * @constructor
 * @param {string} title The title of the note.
 * @param {string} content The content of the note.
 * @param {!Element} noteContainer Container to insert the note into.
 * @export
 */
bugless.notepad.Note = function(title, content, noteContainer) {
    this.title_ = title;
    this.content_ = content;
    this.parent_ = noteContainer;
    /**
     * @type {?HTMLTextAreaElement}
     * @private
     */
    this.editorElement_ = null;
    /**
     * @type {?HTMLDivElement}
     * @private
     */
    this.editorContainer_ = null;
    /**
     * @type {?HTMLDivElement}
     * @private
     */
    this.contentElement_ = null;
};

/**
 * Make DOM for the note.
 */
bugless.notepad.Note.prototype.render = function() {
    let note = goog.dom.createDom(goog.dom.TagName.DIV, null);
    goog.soy.renderElement(note, bugless.templates.bugless.note, {
        'title': this.title_,
        'content': this.content_,
    });

    this.editorElement_ = goog.dom.getElementByTagNameAndClass(goog.dom.TagName.TEXTAREA, "nt", note);
    this.editorContainer_ = goog.dom.getElementByTagNameAndClass(goog.dom.TagName.DIV, "ne", note);
    this.contentElement_ = goog.dom.getElementByTagNameAndClass(goog.dom.TagName.DIV, "nc", note);
    let saveBtn = /** @type {!HTMLInputElement} */ (note.querySelector(".nb"));

    goog.dom.appendChild(this.parent_, note);

    goog.events.listen(this.contentElement_, goog.events.EventType.CLICK, this.openEditor, false, this);
    goog.events.listen(saveBtn, goog.events.EventType.CLICK, this.save, false, this);
};

/**
 * Open note editor.
 */
bugless.notepad.Note.prototype.openEditor = function() {
    this.editorElement_.innerHTML = this.content_;
    this.contentElement_.style.display = "none";
    this.editorContainer_.style.display = "inline";
};

/**
 * Close the note editor.
 */
bugless.notepad.Note.prototype.closeEditor = function() {
    this.contentElement_.innerHTML = this.content_;
    this.contentElement_.style.display = "inline";
    this.editorContainer_.style.display = "none";
};

/**
 * Save the note. Event listener.
 *
 * @param {!goog.events.BrowserEvent} e Event.
 */
bugless.notepad.Note.prototype.save = function(e)  {
    this.content_ = this.editorElement_.value;
    this.closeEditor();
};

/**
 * Makes all notes from list of data.
 *
 * @param {!Array<!{title: string, content: string}>} data Data to insert
 * @param {!Element} noteContainer Container to insert the notes into.
 * @returns {!Array<!bugless.notepad.Note>}
 */
bugless.notepad.makeNotes = (data, noteContainer) => {
    return data.map((datum) => {
        let note = new bugless.notepad.Note(datum.title, datum.content, noteContainer);
        note.render();
        return note;
    });
};

/**
 * Main application entrypoint
 *
 * @param {!Element} container The container to bind the application to.
 * @export
 */
bugless.run = (container) => {
    bugless.notepad.makeNotes([
        { title: "foo", content: "lorem ipsum dolor sit amet?" },
        { title: "baz", content: "lorem ipsum dolor sit amet!" },
    ], container);
};
