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

const KITTY_IMG_HDR = "\x1b_G"
const KITTY_IMG_FTR = "\x1b\\\n"

func IsTermKitty() bool {

	TERM := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	return TERM == "xterm-kitty"
}

func WriteKittyImage(out io.Writer, iImg image.Image) (E error) {

	OSC_OPEN, OSC_CLOSE := KITTY_IMG_HDR, KITTY_IMG_FTR
	if IsTmuxScreen() {
		OSC_OPEN, OSC_CLOSE = TmuxOscOpenClose(OSC_OPEN, OSC_CLOSE)
	}

	B_HDR := []byte(OSC_OPEN)
	B_FTR := []byte(OSC_CLOSE)

	// LAST CHUNK SIGNAL `m=0` TO KITTY
	defer func() {

		if E == nil {
			out.Write(B_HDR)
			out.Write([]byte("m=0;"))
			_, E = out.Write(B_FTR)
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
			B_HDR, b_params, []byte("m=1;"), src, B_FTR,
		}, nil), nil
	}

	enc64 := base64.NewEncoder(base64.StdEncoding, &oWC)
	defer enc64.Close()

	E = png.Encode(enc64, iImg)
	return
}
