package rasterm

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
)

const (
	KITTY_IMG_HDR = "\x1b_G"
	KITTY_IMG_FTR = "\x1b\\"
)

// NOTE: uses $TERM, which is overwritten by tmux
func IsTermKitty() bool {

	V := GetEnvIdentifiers()
	return V["TERM"] == "xterm-kitty"
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

	return S.KittyCopyPNGInline(out, pBuf, int64(pBuf.Len()))
}

// Encode raw PNG data into Kitty terminal format
func (S Settings) KittyCopyPNGInline(out io.Writer, in io.Reader, nLen int64) (E error) {

	// OPTIONALLY TMUX-ESCAPE OPENING & CLOSING OSC CODES
	OSC_OPEN, OSC_CLOSE := KITTY_IMG_HDR, KITTY_IMG_FTR
	if S.EscapeTmux {
		OSC_OPEN, OSC_CLOSE = TmuxOscOpenClose(OSC_OPEN, OSC_CLOSE)
	}

	// LAST CHUNK SIGNAL `m=0` TO KITTY
	defer func() {

		if E == nil {
			out.Write([]byte(OSC_OPEN))
			out.Write([]byte("m=0;"))
			_, E = out.Write([]byte(OSC_CLOSE))
		}
	}()

	// PIPELINE: PNG -> B64 -> CHUNKER -> out io.Writer
	// SEND IN 4K CHUNKS
	oWC := NewWriteChunker(out, 4096)
	defer oWC.Flush()
	bsHdr := []byte(fmt.Sprintf("a=T,f=100,z=-1,S=%d,", nLen))
	oWC.CustomWriFunc = func(iWri io.Writer, bsDat []byte) (int, error) {

		parts := [][]byte{
			[]byte(OSC_OPEN),
			bsHdr,
			[]byte("m=1;"),
			bsDat,
			[]byte(OSC_CLOSE),
		}

		bsHdr = nil

		return iWri.Write(bytes.Join(parts, nil))
	}

	enc64 := base64.NewEncoder(base64.StdEncoding, &oWC)
	defer enc64.Close()

	_, E = io.Copy(enc64, in)
	return
}
