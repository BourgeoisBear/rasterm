package rasterm

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
)

// See https://sw.kovidgoyal.net/kitty/graphics-protocol.html for more details.

const (
	KITTY_IMG_HDR = "\x1b_G"
	KITTY_IMG_FTR = "\x1b\\"
)

func IsTermKitty() bool {

	V := GetEnvIdentifiers()
	return len(V["KITTY_WINDOW_ID"]) > 0
}

// Display local PNG file
// - pngFileName must be directly accesssible from Kitty instance
// - pngFileName must be an absolute path
func (S Settings) KittyWritePNGLocal(out io.Writer, pngFileName string) error {

	_, e := fmt.Fprint(out, KITTY_IMG_HDR, "a=T,f=100,t=f;")
	if e != nil {
		return e
	}

	enc64 := base64.NewEncoder(base64.RawStdEncoding, out)

	_, e = fmt.Fprint(enc64, pngFileName)
	if e != nil {
		return e
	}

	e = enc64.Close()
	if e != nil {
		return e
	}

	_, e = fmt.Fprint(out, KITTY_IMG_FTR)
	return e
}

// Serialize image.Image into Kitty terminal in-band format.
func (S Settings) KittyWriteImage(out io.Writer, iImg image.Image) error {

	pBuf := new(bytes.Buffer)
	if E := png.Encode(pBuf, iImg); E != nil {
		return E
	}

	return S.KittyCopyPNGInline(out, pBuf)
}

// Serialize PNG image from io.Reader into Kitty terminal in-band format.
func (S Settings) KittyCopyPNGInline(out io.Writer, in io.Reader) error {

	// PIPELINE: PNG (io.Reader) -> B64 -> CHUNKER -> (io.Writer)
	// SEND IN 4K CHUNKS
	cw := kittyChunkWri{
		nChunkSize: 4096,
		iWri:       out,
	}

	enc64 := base64.NewEncoder(base64.RawStdEncoding, &cw)
	_, err := io.Copy(enc64, in)
	return errors.Join(
		err,
		enc64.Close(),
		cw.Close(),
	)
}
