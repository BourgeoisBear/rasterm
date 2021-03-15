package rasterm

import "io"

/*
Used by WriteChunker to optionally transform chunks before
sending them on to the underlying io.Writer.
*/
type XfrmFunc func([]byte) ([]byte, error)

/*
Wraps an io.Writer interface to buffer/flush in chunks that are
`chunkSize` bytes long.  Optional `Xfrm` function in struct
allows for additional []byte processing before sending each
chunk to the underlying writer. Currently used for encoding to
Kitty terminal's image format.
*/
type WriteChunker struct {
	chunk  []byte
	writer io.Writer
	ix     int
	Xfrm   XfrmFunc
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

func (pC *WriteChunker) Flush() (E error) {

	tmp := pC.chunk[:pC.ix]
	if pC.Xfrm != nil {
		if tmp, E = pC.Xfrm(tmp); E != nil {
			return
		}
	}

	_, E = pC.writer.Write(tmp)
	pC.ix = 0
	return
}

func (pC *WriteChunker) Write(src []byte) (int, error) {

	chunkSize := len(pC.chunk)

	for _, bt := range src {

		pC.chunk[pC.ix] = bt
		pC.ix++

		if pC.ix >= chunkSize {
			if e := pC.Flush(); e != nil {
				return 0, e
			}
		}
	}

	return len(src), nil
}
