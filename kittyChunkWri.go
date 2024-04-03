package rasterm

import (
	"fmt"
	"io"
)

type kittyChunkWri struct {
	iWri       io.Writer
	nChunkSize int
	nMod       int
}

// signal last chunk to Kitty on close
func (c *kittyChunkWri) Close() error {
	_, err := fmt.Fprint(c.iWri, KITTY_IMG_HDR, "m=0;", KITTY_IMG_FTR)
	return err
}

func (c *kittyChunkWri) Write(buf []byte) (int, error) {

	l := len(buf)
	var toWrite, nWritten int

	for l > 0 {

		if (c.nMod + l) >= c.nChunkSize {
			toWrite = c.nChunkSize - c.nMod
			c.nMod = 0
		} else {
			toWrite = l
			c.nMod += l
		}

		// prefix
		_, err := fmt.Fprint(c.iWri, KITTY_IMG_HDR, "m=1;")
		if err != nil {
			return nWritten, err
		}

		// data
		var n int
		n, err = c.iWri.Write(buf[:toWrite])
		nWritten += n
		if err != nil {
			return nWritten, err
		}

		// suffix
		_, err = fmt.Fprint(c.iWri, KITTY_IMG_FTR)
		if err != nil {
			return nWritten, err
		}

		buf = buf[toWrite:]
		l -= toWrite
	}

	return nWritten, nil
}
