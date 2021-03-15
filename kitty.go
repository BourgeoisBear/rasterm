package rasterm

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"io"
	"os"
	"strings"
)

const (
	KITTY_IMG_HDR = "\x1b_G"
	KITTY_IMG_FTR = "\x1b\\\n"
)

// NOTE: uses $TERM, which is overwritten by tmux
func IsTermKitty() bool {

	TERM := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	return TERM == "xterm-kitty"
}

/*
Encode image using the Kitty terminal graphics protocol: https://sw.kovidgoyal.net/kitty/graphics-protocol.html
*/
func (S Settings) WriteKittyImage(out io.Writer, iImg image.Image) (E error) {

	OSC_OPEN, OSC_CLOSE := KITTY_IMG_HDR, KITTY_IMG_FTR
	if S.EscapeTmux && IsTmuxScreen() {
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
	var b_params []byte
	oWC.Xfrm = func(src []byte) ([]byte, error) {

		// os.Stderr.Write([]byte(fmt.Sprintf("%d", len(src))))
		if b_params == nil {
			b_params = []byte("a=T,f=100,z=-1,")
		} else {
			b_params = nil
		}

		return bytes.Join([][]byte{
			[]byte(OSC_OPEN), b_params, []byte("m=1;"), src, []byte(OSC_CLOSE),
		}, nil), nil
	}

	enc64 := base64.NewEncoder(base64.StdEncoding, &oWC)
	defer enc64.Close()

	E = png.Encode(enc64, iImg)
	return
}
