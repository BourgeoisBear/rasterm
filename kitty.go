package rasterm

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"io"
)

const (
	KITTY_IMG_HDR = "\x1b_G"
	KITTY_IMG_FTR = "\x1b\\"
)

func IsTermKitty() bool {

	V := GetEnvIdentifiers()
	return len(V["KITTY_WINDOW_ID"]) > 0
}

/*
Encode image using the Kitty terminal graphics protocol:
https://sw.kovidgoyal.net/kitty/graphics-protocol.html
*/
func (S Settings) KittyWriteImage(out io.Writer, iImg image.Image) error {

	pBuf := new(bytes.Buffer)
	if E := png.Encode(pBuf, iImg); E != nil {
		return E
	}

	return S.KittyCopyPNGInline(out, pBuf)
}

// Encode raw PNG data into Kitty terminal format
func (S Settings) KittyCopyPNGInline(out io.Writer, in io.Reader) (E error) {

	// OPTIONALLY TMUX-ESCAPE OPENING & CLOSING OSC CODES
	OSC_OPEN, OSC_CLOSE := KITTY_IMG_HDR, KITTY_IMG_FTR

	// LAST CHUNK SIGNAL `m=0` TO KITTY
	defer func() {

		if E == nil {
			_, E = writeMulti(out, [][]byte{
				[]byte(OSC_OPEN),
				[]byte("m=0;"),
				[]byte(OSC_CLOSE),
			})
		}
	}()

	// PIPELINE: PNG -> B64 -> CHUNKER -> out io.Writer
	// SEND IN 4K CHUNKS
	oWC := NewWriteChunker(out, 4096)
	defer oWC.Flush()
	bsHdr := []byte("a=T,f=100,")
	oWC.CustomWriFunc = func(iWri io.Writer, bsDat []byte) (int, error) {

		parts := [][]byte{
			[]byte(OSC_OPEN),
			bsHdr,
			[]byte("m=1;"),
			bsDat,
			[]byte(OSC_CLOSE),
		}

		bsHdr = nil

		return writeMulti(iWri, parts)
	}

	enc64 := base64.NewEncoder(base64.StdEncoding, &oWC)
	defer enc64.Close()

	_, E = io.Copy(enc64, in)
	return
}

func writeMulti(iWri io.Writer, parts [][]byte) (int, error) {

	var nTot int
	for ix := range parts {
		if len(parts[ix]) == 0 {
			continue
		}
		n, ew := iWri.Write(parts[ix])
		nTot += n
		if ew != nil {
			return nTot, ew
		}
	}
	return nTot, nil
}
