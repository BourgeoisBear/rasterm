package rasterm

import (
	"fmt"
	"image"
	"io"
	"strconv"
)

/*
	Encodes a paletted image into SIXEL format.
	Forked & heavily modified from https://github.com/mattn/go-sixel/
*/
type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

const (
	SIXEL_MIN byte = 0x3f
	SIXEL_MAX byte = 0x7e
)

/*
	TODO:
		- documentation
		- better terminal detection
		- write output-mode selection multiplexer
		- unit tests for larger photos
*/

/*
	NOTE:
		this does not support transparency.
		alpha values in the palette will be ignored.

	encodes paletted image into DECSIXEL format,
	outputs to underlying io.Writer

	For more information on DECSIXEL format:
		https://www.vt100.net/docs/vt3xx-gp/chapter14.html
*/
func (e *Encoder) Encode(pI *image.Paletted) (E error) {

	width, height := pI.Bounds().Dx(), pI.Bounds().Dy()
	if (width <= 0) || (height <= 0) {
		return
	}

	if len(pI.Palette) == 0 {
		return
	}

	// CAPTURE WRITE ERROR FOR SIMPLIFIED CHECKING
	fnWri := func(v []byte) error {
		_, E = e.w.Write(v)
		return E
	}

	// TMUX/SCREEN WORKAROUND
	OSC_OPEN, OSC_CLOSE := "\x1b", "\x1b\\"
	if IsTmuxScreen() {
		OSC_OPEN, OSC_CLOSE = TmuxOscOpenClose(OSC_OPEN, OSC_CLOSE)
	}

	// INTRODUCER = <ESC>P0;1q
	// 0; rely on RASTER ATTRIBUTES to set aspect ratio
	// 1; palette[0] as opaque
	// RASTER ATTRIBUTES (1:1 aspect ratio) = "1;1;width;height
	_, E = fmt.Fprintf(e.w, "%sP0;1q\"1;1;%d;%d\n", OSC_OPEN, width, height)
	if E != nil {
		return
	}

	// CONVERT uint32 [0..0xFFFF] COLOR COMPONENT TO WHOLE PERCENTAGE
	P := func(v uint32) uint8 {
		return uint8(((v + 1) * 100) >> 16)
	}

	// SEND PALETTE
	for ix_color, v := range pI.Palette {

		// SIXEL ONLY SUPPORTS 256 COLORS
		if ix_color > 255 {
			break
		}

		// R,G,B AS WHOLE PERCENTAGES
		r, g, b, _ := v.RGBA()

		// DECGCI (#): Graphics Color Introducer
		// SEE: https://www.vt100.net/docs/vt3xx-gp/chapter14.html
		_, E = fmt.Fprintf(e.w, "#%d;2;%d;%d;%d", ix_color, P(r), P(g), P(b))
		if E != nil {
			return
		}
	}

	nColors := len(pI.Palette)
	color_used := make([]bool, nColors)
	color_used_blank := make([]bool, nColors)
	buf := make([]byte, width*nColors)
	buf_blank := make([]byte, width*nColors)

	// WALK IMAGE HEIGHT IN SIXEL ROWS
	sixel_rows := (height + 5) / 6
	for ix_srow := 0; ix_srow < sixel_rows; ix_srow++ {

		// GRAPHICS NL (start a new sixel line)
		if ix_srow > 0 {
			if fnWri([]byte(`-`)) != nil {
				return
			}
		}

		// RESET COLOR USAGE FLAGS
		copy(color_used, color_used_blank)
		copy(buf, buf_blank)

		// BUFFER SIXEL ROW, TRACK USED COLORS
		for p := 0; p < 6; p++ {

			y := (ix_srow * 6) + p
			for x := 0; x < width; x++ {
				color_ix := pI.ColorIndexAt(x, y)
				color_used[color_ix] = true
				buf[(width*int(color_ix))+x] |= 1 << uint(p)
			}
		}

		// RENDER SIXEL ROW FOR EACH PALETTE ENTRY
		bFirstColorWritten := false
		for n := 0; n < nColors; n++ {

			if !color_used[n] {
				continue
			}

			// GRAPHICS CR (overwrite last line w/ new color)
			if bFirstColorWritten {
				if fnWri([]byte(`$`)) != nil {
					return
				}
			}

			// COLOR INTRODUCER (#)
			tmpCI := make([]byte, 1, 4)
			tmpCI[0] = byte('#')
			tmpCI = strconv.AppendInt(tmpCI, int64(n), 10)
			if fnWri(tmpCI) != nil {
				return
			}

			rleCt := 0
			cPrev := byte(255)
			for x := 0; x < width; x++ {

				// GET BUFFERED SIXEL, CLEAR BUFFER
				cNext := buf[(n*width)+x]

				// RLE ENCODE, WRITE ON VALUE CHANGE
				// USE 255 AS SENTINEL FOR INITIAL RUN
				if (cPrev != 255) && (cNext != cPrev) {

					if fnWri(encodeGRI(rleCt, cPrev)) != nil {
						return
					}
					rleCt = 0
				}

				cPrev = cNext
				rleCt++
			}

			// WRITE LAST SIXEL IN LINE
			if fnWri(encodeGRI(rleCt, cPrev)) != nil {
				return
			}

			bFirstColorWritten = true
		}
	}

	// SIXEL TERMINATOR
	fnWri([]byte(OSC_CLOSE))
	return
}

/*
	NOTE: valid sixel encodeds are in range
		0x3F (?) TO 0x7E (~)
*/
func encodeGRI(rleCt int, sixl byte) []byte {

	if rleCt <= 0 {
		return nil
	}

	// MASK WITH VALID SIXEL BITS, APPLY OFFSET
	sixl = SIXEL_MIN + (sixl & 0b111111)
	tmpGRI := make([]byte, 0, 6)

	if rleCt > 3 {

		// GRAPHICS REPEAT INTRODUCER (!<repeat count><sixel>)
		tmpGRI = append(tmpGRI, byte('!'))
		tmpGRI = strconv.AppendInt(tmpGRI, int64(rleCt), 10)
		tmpGRI = append(tmpGRI, sixl)

	} else if rleCt > 0 {

		for ix := 0; ix < rleCt; ix++ {
			tmpGRI = append(tmpGRI, sixl)
		}
	}

	return tmpGRI
}
