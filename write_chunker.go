package rasterm

import "io"

/*
Used by WriteChunker to optionally transform chunks before
sending them on to the underlying io.Writer.
*/
type CustomWriFunc func(io.Writer, []byte) (int, error)

/*
Wraps an io.Writer interface to buffer/flush in chunks that are
`chunkSize` bytes long.  Optional `CustomWriFunc` in struct
allows for additional []byte processing before sending each
chunk to the underlying writer. Currently used for encoding to
Kitty terminal's image format.
*/
type WriteChunker struct {
	chunk  []byte
	writer io.Writer
	ix     int
	CustomWriFunc
}

func NewWriteChunker(iWri io.Writer, chunkSize int) WriteChunker {

	if chunkSize < 1 {
		panic("invalid chunk size")
	}

	return WriteChunker{
		chunk:  make([]byte, chunkSize),
		writer: iWri,
	}
}

func (pC *WriteChunker) Flush() (int, error) {

	tmp := pC.chunk[:pC.ix]
	pC.ix = 0
	if pC.CustomWriFunc != nil {
		return pC.CustomWriFunc(pC.writer, tmp)
	}
	return pC.writer.Write(tmp)
}

func (pC *WriteChunker) Write(src []byte) (int, error) {

	var nTot int
	chunkSize := len(pC.chunk)

	for _, bt := range src {

		pC.chunk[pC.ix] = bt
		pC.ix++

		if pC.ix >= chunkSize {
			n, e := pC.Flush()
			nTot += n
			if e != nil {
				return nTot, e
			}
		}
	}

	return len(src), nil
}
