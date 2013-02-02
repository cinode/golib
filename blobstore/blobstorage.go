package blobstore

import (
	"errors"
	"io"
)

var (
	ErrBIDCollision = errors.New("A colliding BID has been found")
	ErrBIDNotFound  = errors.New("A blob with given BID was not found")
)

type BlobWriter interface {
	io.Writer

	// Finalize blob generation, if no error is returned,
	// the duplicate flag will indicate whether this blob
	// was already inside the blobstore and is equal to the
	// new one written
	Finalize() error

	// Cancel the blob generation
	Cancel() error
}

// An interface usefull for blob storage operations
type BlobStorage interface {

	// Create new writer for blobs
	NewBlobWriter(blobId string) (writer BlobWriter, err error)

	// Create new reader for existing blob
	NewBlobReader(blobId string) (reader io.Reader, err error)
}
