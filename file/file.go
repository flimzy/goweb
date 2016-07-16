// +build js

// File provides information about files and allows JavaScript in a web page to access their content.
//
// See https://developer.mozilla.org/en-US/docs/Web/API/File
package file

import (
	"time"

	"github.com/flimzy/goweb/blob"
	"github.com/gopherjs/gopherjs/js"
)

type File struct {
	blob.Blob
	// Returns the name of the file referenced by the File object. Read only.
	Name string `js:"name"`
}

func Internalize(o *js.Object) *File {
	return &File{Blob: *blob.Internalize(o)}
}

// Returns a time.Time struct representing the last modified time of the file.
func (f *File) LastModified() time.Time {
	ts := f.Get("lastModified").Int64()
	return time.Unix(ts/1000, (ts%1000)*1000000)
}
