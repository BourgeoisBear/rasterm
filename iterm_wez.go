package rasterm

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"strings"
)

const (
	ITERM_IMG_HDR = "\x1b]1337;File=inline=1"
	ITERM_IMG_FTR = "\a"
)

// NOTE: uses $TERM_PROGRAM, which isn't passed through tmux
func IsTermItermWez() bool {

	TERM_PROG := strings.ToLower(strings.TrimSpace(os.Getenv("TERM_PROGRAM")))

	switch TERM_PROG {
	case "iterm.app", "wezterm":
		return true
	}

	return false
}

/*
Encode image using the iTerm2/WezTerm terminal image protocol:
https://iterm2.com/documentation-images.html
*/
func (S Settings) ItermWriteImage(out io.Writer, iImg image.Image) error {

	pBuf := new(bytes.Buffer)
	if E := png.Encode(pBuf, iImg); E != nil {
		return E
	}

	return S.ItermCopyFileInline(out, pBuf, int64(pBuf.Len()))
}

func (S Settings) ItermCopyFileInline(out io.Writer, in io.Reader, nLen int64) (E error) {

	OSC_OPEN, OSC_CLOSE := ITERM_IMG_HDR, ITERM_IMG_FTR
	if S.EscapeTmux && IsTmuxScreen() {
		OSC_OPEN, OSC_CLOSE = TmuxOscOpenClose(OSC_OPEN, OSC_CLOSE)
	}

	if _, E = out.Write([]byte(OSC_OPEN)); E != nil {
		return
	}

	hdrSize := fmt.Sprintf(";size=%d:", nLen)
	if _, E = out.Write([]byte(hdrSize)); E != nil {
		return
	}

	enc64 := base64.NewEncoder(base64.StdEncoding, out)
	if _, E = io.Copy(enc64, in); E != nil {
		return
	}

	if E = enc64.Close(); E != nil {
		return
	}

	_, E = out.Write([]byte(OSC_CLOSE))
	return
}
