package rasterm

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"strconv"
	"strings"
)

const (
	ITERM_IMG_HDR = "\x1b]1337;File="
	ITERM_IMG_FTR = "\a"
)

type ItermImgOpts struct {
	// Filename. Defaults to "Unnamed file".
	Name string

	// Width to render. See notes below.
	Width string

	// Height to render. See notes below.
	Height string

	// The width and height are given as a number followed by a unit, or the word "auto".
	//
	//   - N: N character cells.
	//   - Npx: N pixels.
	//   - N%: N percent of the session's width or height.
	//   - auto: The image's inherent size will be used to determine an appropriate dimension.

	// File size in bytes. Optional; this is only used by the progress indicator.
	Size int64

	// If set, the file will be displayed inline. Otherwise, it will be downloaded
	// with no visual representation in the terminal session.
	DisplayInline bool

	// If set, the image's inherent aspect ratio will not be respected.
	IgnoreAspectRatio bool
}

func (o ItermImgOpts) ToHeader() string {

	var opts []string

	if o.Name != "" {
		opts = append(opts, "name="+base64.StdEncoding.EncodeToString([]byte(o.Name)))
	}

	if o.Width != "" {
		opts = append(opts, "width="+o.Width)
	}

	if o.Height != "" {
		opts = append(opts, "height="+o.Height)
	}

	if o.Size > 0 {
		opts = append(opts, "size="+strconv.FormatInt(o.Size, 10))
	}

	// default: inline=0
	if o.DisplayInline {
		opts = append(opts, "inline=1")
	}

	// default: preserveAspectRatio=1
	if o.IgnoreAspectRatio {
		opts = append(opts, "preserveAspectRatio=0")
	}

	return ITERM_IMG_HDR + strings.Join(opts, ";") + ":"
}

// NOTE: uses $TERM_PROGRAM, which isn't passed through tmux or ssh
// checks if iterm inline image protocol is supported
func IsItermCapable() bool {

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

	if V["TERM_PROGRAM"] == "rio" {
		return true
	}

	return false
}

/*
Encode image using the iTerm2/WezTerm terminal image protocol:

	https://iterm2.com/documentation-images.html
*/
func ItermWriteImageWithOptions(out io.Writer, iImg image.Image, opts ItermImgOpts) error {

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

	opts.Size = int64(pBuf.Len())
	return ItermCopyFileInlineWithOptions(out, pBuf, opts)
}

func ItermCopyFileInlineWithOptions(out io.Writer, in io.Reader, opts ItermImgOpts) (E error) {

	if _, E = fmt.Fprint(out, opts.ToHeader()); E != nil {
		return
	}

	enc64 := base64.NewEncoder(base64.StdEncoding, out)
	if _, E = io.Copy(enc64, in); E != nil {
		return
	}

	if E = enc64.Close(); E != nil {
		return
	}

	_, E = out.Write([]byte(ITERM_IMG_FTR))
	return
}

func ItermWriteImage(out io.Writer, iImg image.Image) error {
	return ItermWriteImageWithOptions(out, iImg, ItermImgOpts{DisplayInline: true})
}

func ItermCopyFileInline(out io.Writer, in io.Reader, nLen int64) (E error) {
	return ItermCopyFileInlineWithOptions(out, in, ItermImgOpts{DisplayInline: true, Size: nLen})
}
