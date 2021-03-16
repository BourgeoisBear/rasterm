package rasterm

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"testing"
)

var testFiles []string

func init() {

	files, e := os.ReadDir("./test_images")
	if e != nil {
		panic(e)
	}

	for ix := range files {
		testFiles = append(testFiles, files[ix].Name())
	}

	os.Stdout.Write([]byte(ESC_ERASE_DISPLAY))
}

func loadImage(path string) (iImg image.Image, imgFmt string, E error) {

	pF, E := os.Open(path)
	if E != nil {
		return
	}
	defer pF.Close()

	return image.Decode(pF)
}

func getFile(fpath string) (*os.File, int64, error) {

	pF, E := os.Open(fpath)
	if E != nil {
		return nil, 0, E
	}

	fInf, E := pF.Stat()
	if E != nil {
		pF.Close()
		return nil, 0, E
	}

	return pF, fInf.Size(), nil
}

func getImgInfo(pF *os.File) (imgCfg image.Config, fmtName string, E error) {

	if imgCfg, fmtName, E = image.DecodeConfig(pF); E != nil {
		return
	}

	// REWIND FILE
	_, E = pF.Seek(0, 0)
	return
}

type TestLogger interface {
	Log(...interface{})
	Logf(string, ...interface{})
}

func testEx(iLog TestLogger, out io.Writer, mode string, testFiles []string) error {

	/*
		fProf, E := os.Create("./kitty.prof")
		if E != nil {
			return E
		}
		defer fProf.Close()
		pprof.StartCPUProfile(fProf)
		defer pprof.StopCPUProfile()
	*/

	S := Settings{
		EscapeTmux: false,
	}

	for _, file := range testFiles {

		fpath := "./test_images/" + file
		iLog.Log(fpath)

		fIn, nImgLen, e2 := getFile(fpath)
		if e2 != nil {
			iLog.Log(e2)
			continue
		}
		defer fIn.Close()

		iLog.Logf("IMAGE SIZE %d", nImgLen)

		imgCfg, fmtName, e2 := getImgInfo(fIn)
		if e2 != nil {
			iLog.Log(e2)
			continue
		}

		iLog.Logf("FMT: %s, W: %d, H: %d", fmtName, imgCfg.Width, imgCfg.Height)

		iImg, _, e2 := loadImage(fpath)
		if e2 != nil {
			iLog.Log(e2)
			continue
		}

		var e3 error = nil
		switch mode {
		case "iterm":

			// e3 = S.ItermWriteImage(out, iImg)

			// WEZ/ITERM SUPPORT ALL FORMATS, SO NO NEED TO RE-ENCODE TO PNG
			e3 = S.ItermCopyFileInline(out, fIn, nImgLen)

		case "sixel":

			if iPaletted, bOK := iImg.(*image.Paletted); bOK {

				e3 = S.SixelWriteImage(out, iPaletted)

			} else {

				iLog.Logf("%s is type [%T], not *image.Paletted\n", file, iImg)
				continue
			}

		case "kitty":

			if fmtName == "png" {

				e3 = S.KittyCopyPNGInline(out, fIn, nImgLen)

			} else {

				e3 = S.KittyWriteImage(out, iImg)
			}
		}

		if e3 != nil {
			iLog.Log(e3)
		}
		fmt.Println("")
	}

	return nil
}

// NOTE
//
// can't query terminal attributes here (i.e. sixel support) since golang
// testbed intermediates itself between stdin/stdout with buffers
func TestSixel(pT *testing.T) {

	if IsTermItermWez() || IsTermKitty() {
		pT.SkipNow()
	}

	if E := testEx(pT, os.Stdout, "sixel", testFiles); E != nil {
		pT.Fatal(E)
	}
}

func TestItermWez(pT *testing.T) {

	if !IsTermItermWez() {
		pT.SkipNow()
	}

	if E := testEx(pT, os.Stdout, "iterm", testFiles); E != nil {
		pT.Fatal(E)
	}
}

func TestKitty(pT *testing.T) {

	if !IsTermKitty() {
		pT.SkipNow()
	}

	if E := testEx(pT, os.Stdout, "kitty", testFiles); E != nil {
		pT.Fatal(E)
	}
}
