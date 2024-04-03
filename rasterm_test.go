package rasterm

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
)

var testFiles []string

func init() {

	files, e := os.ReadDir("./test_images")
	if e != nil {
		panic(e)
	}

	for _, fi := range files {
		switch fi.Name() {
		default:
			testFiles = append(testFiles, fi.Name())
		}
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

	baseDir, e2 := filepath.Abs("./test_images")
	if e2 != nil {
		return e2
	}

	for _, file := range testFiles {

		fpath := baseDir + "/" + file
		iLog.Log(fpath)

		fIn, nImgLen, e2 := getFile(fpath)
		if e2 != nil {
			iLog.Log(e2)
			continue
		}
		defer fIn.Close()

		fmt.Println(fpath)

		imgCfg, fmtName, e2 := getImgInfo(fIn)
		if e2 != nil {
			iLog.Log(e2)
			continue
		}

		iImg, _, e2 := loadImage(fpath)
		if e2 != nil {
			iLog.Log(e2)
			continue
		}

		fmt.Printf("[FMT: %s, W: %d, H: %d, LEN: %d, IMG: %T]\n", fmtName, imgCfg.Width, imgCfg.Height, nImgLen, iImg)

		var e3 error = nil
		switch mode {
		case "iterm":

			// WEZ/ITERM SUPPORT ALL FORMATS, SO NO NEED TO RE-ENCODE TO PNG
			e3 = ItermCopyFileInline(out, fIn, nImgLen)

		case "sixel":

			if iPaletted, bOK := iImg.(*image.Paletted); bOK {

				e3 = SixelWriteImage(out, iPaletted)

			} else {

				fmt.Println("[NOT PALETTED, SKIPPING.]")
				continue
			}

		case "kitty":

			if fmtName == "png" {

				fmt.Println("Kitty PNG Local File")
				eF := KittyWritePNGLocal(out, fpath, KittyImgOpts{})
				fmt.Println("\nKitty PNG Inline")
				eI := KittyCopyPNGInline(out, fIn, KittyImgOpts{})
				e3 = errors.Join(eI, eF)

			} else {

				e3 = KittyWriteImage(out, iImg, KittyImgOpts{})
			}
		}

		if e3 != nil {
			iLog.Log(e3)
		}
		fmt.Println("")
	}

	return nil
}

func TestSixel(pT *testing.T) {

	// NOTE: go test captures stdin/stdout to where they are no longer TTYs.
	// This prevents IsSixelCapable() from operating, so always attempting
	// sixels from the test, whether the terminal is capable or not.
	// https://github.com/golang/go/issues/18153

	// bSix, err := IsSixelCapable()
	// if err != nil {
	// 	pT.Fatal(err)
	// }

	fmt.Println("SIXEL")
	if E := testEx(pT, os.Stdout, "sixel", testFiles); E != nil {
		pT.Fatal(E)
	}
}

func TestItermWez(pT *testing.T) {

	if !IsItermCapable() {
		pT.SkipNow()
	}

	fmt.Println("ITERM/WEZ/MINTTY")
	if E := testEx(pT, os.Stdout, "iterm", testFiles); E != nil {
		pT.Fatal(E)
	}
}

func TestKitty(pT *testing.T) {

	if !IsKittyCapable() {
		pT.SkipNow()
	}

	fmt.Println("KITTY")
	if E := testEx(pT, os.Stdout, "kitty", testFiles); E != nil {
		pT.Fatal(E)
	}
}
