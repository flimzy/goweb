// +build js

package blob

import (
	"bytes"
	"errors"
	// 	"sync"

	"github.com/flimzy/goweb/event"
	"github.com/gopherjs/gopherjs/js"
)

// Blob wraps a js.Object
type Blob struct {
	js.Object
	// A boolean value, indicating whether the Blob.close() method has been called on the blob. Closed blobs can not be read. (Read-only)
	IsClosed bool `js:"isClosed"`
	// The size, in bytes, of the data contained in the Blob object. (Read-only)
	Size int64 `js:"size"`
	// A string indicating the MIME type of the data contained in the Blob. If the type is unknown, this string is empty. (Read-only)
	Type string `js:"type"`
}

type Options struct {
	js.Object
	Type    string `js:"type"`
	Endings string `js:"endings"`
}

// New returns a newly created Blob object whose content consists of the
// concatenation of the array of values given in parameter.
func New(parts []interface{}, opts Options) *Blob {
	blob := js.Global.Get("Blob").New(parts, opts)
	return &Blob{Object: *blob}
}

// Internalize internalizes a standard *js.Object to a GlobObj
func Internalize(o *js.Object) *Blob {
	return &Blob{Object: *o}
}

// Close closes the blob object, possibly freeing underlying resources.
func (b *Blob) Close() {
	b.Call("close")
}

// Slice returns a new Blob object containing the specified range of bytes of the source BlobObject.
func (b *Blob) Slice(start, end int, contenttype string) *Blob {
	newBlobObject := b.Call("slice", start, end, contenttype)
	return &Blob{
		Object: *newBlobObject,
	}
}

type Reader struct {
	r    *bytes.Reader
	err  error
	done <-chan struct{}
	fr   *js.Object
}

// NewReader returns a new Reader reading from b.
func (b *Blob) NewReader() *Reader {
	done := make(chan struct{})
	fr := js.Global.Get("FileReader").New()
	r := &Reader{
		done: done,
		fr:   fr,
	}
	handler := func(e *event.Event) {
		r.fr = nil // Remove the FileReader from memory, now that it's no longer in use
		defer close(done)
		if err := fr.Get("error"); err != nil {
			r.err = &js.Error{err}
		} else if fr.Get("readyState").Int() != 2 {
			r.err = errors.New("Unexpected readyState")
		}
		content := js.Global.Get("Uint8Array").New(fr.Get("result")).Interface().([]byte)
		r.r = bytes.NewReader(content)
	}
	event.AddEventListener(fr, "loaded", handler, event.ListenerOpts{})
	event.AddEventListener(fr, "onerror", handler, event.ListenerOpts{})
	event.AddEventListener(fr, "onabort", handler, event.ListenerOpts{})
	fr.Call("readAsArrayBuffer", b)
	return r
}

// Abort aborts a running read operation.
//
// See https://developer.mozilla.org/en-US/docs/Web/API/FileReader/abort
func (r *Reader) Abort() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(r.(string))
		}
	}()
	if r.fr != nil {
		r.fr.Call("abort")
		return
	}
	return errors.New("Read not in progress")
}

// Read implements the io.Reader interface
func (r *Reader) Read(p []byte) (int, error) {
	<-r.done // Make sure we're done reading
	if r.err != nil {
		return 0, r.err
	}
	return r.r.Read(p)
}
