package rasterm

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

const (
	ITERM_IMG_HDR = "\x1b]1337;File=inline=1"
	ITERM_IMG_FTR = "\a"
)

// NOTE: uses $TERM_PROGRAM, which isn't passed through tmux or ssh
func IsTermItermWez() bool {

	V := GetEnvIdentifiers()

	if V["TERM"] == "mintty" {
		return true
	}

	if V["LC_TERMINAL"] == "iterm2" {
		return true
	}

	if V["TERM_PROGRAM"] == "wezterm" {
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
	var E error

	// NOTE: doing this under suspicion that wezterm PNG handling is slow
	if _, bOK := iImg.(*image.Paletted); bOK {

		// PNG IF PALETTED
		E = png.Encode(pBuf, iImg)

	} else {

		// JPG IF NOT
		E = jpeg.Encode(pBuf, iImg, &jpeg.Options{Quality: 93})
	}

	if E != nil {
		return E
	}

	return S.ItermCopyFileInline(out, pBuf, int64(pBuf.Len()))
}

func (S Settings) ItermCopyFileInline(out io.Writer, in io.Reader, nLen int64) (E error) {

	OSC_OPEN, OSC_CLOSE := ITERM_IMG_HDR, ITERM_IMG_FTR

	if _, E = out.Write([]byte(OSC_OPEN)); E != nil {
		return
	}

	if _, E = fmt.Fprintf(out, ";size=%d:", nLen); E != nil {
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
