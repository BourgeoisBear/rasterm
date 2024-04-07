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

type TestLogger interface {
	Log(...interface{})
	Logf(string, ...interface{})
}

func testImage(iWri io.Writer, fpath, mode string) error {

	fIn, nImgLen, err := getFile(fpath)
	if err != nil {
		return err
	}
	defer fIn.Close()

	fmt.Println(fpath)

	imgCfg, fmtName, err := image.DecodeConfig(fIn)
	if err != nil {
		return err
	}

	_, err = fIn.Seek(0, 0)
	if err != nil {
		return err
	}

	iImg, _, err := image.Decode(fIn)
	if err != nil {
		return err
	}

	_, err = fIn.Seek(0, 0)
	if err != nil {
		return err
	}

	fmt.Printf("[FMT: %s, W: %d, H: %d, LEN: %d, IMG: %T]\n", fmtName, imgCfg.Width, imgCfg.Height, nImgLen, iImg)

	switch mode {
	case "iterm":

		// WEZ/ITERM SUPPORT ALL FORMATS, SO NO NEED TO RE-ENCODE TO PNG
		err = ItermCopyFileInline(iWri, fIn, nImgLen)

	case "sixel":

		if iPaletted, bOK := iImg.(*image.Paletted); bOK {

			err = SixelWriteImage(iWri, iPaletted)

		} else {

			fmt.Println("[NOT PALETTED, SKIPPING.]")
		}

	case "kitty":

		if fmtName == "png" {

			fmt.Println("Kitty PNG Local File")
			eF := KittyWritePNGLocal(iWri, fpath, KittyImgOpts{})
			fmt.Println("\nKitty PNG Inline")
			eI := KittyCopyPNGInline(iWri, fIn, KittyImgOpts{})
			err = errors.Join(eI, eF)

		} else {

			err = KittyWriteImage(iWri, iImg, KittyImgOpts{})
		}
	}

	fmt.Println("")
	return err
}

func testEx(iLog TestLogger, iWri io.Writer, mode string, testFiles []string) error {

	/*
		fProf, E := os.Create("./kitty.prof")
		if E != nil {
			return E
		}
		defer fProf.Close()
		pprof.StartCPUProfile(fProf)
		defer pprof.StopCPUProfile()
	*/

	baseDir, err := filepath.Abs("./test_images")
	if err != nil {
		return err
	}

	for _, file := range testFiles {

		fpath := baseDir + "/" + file
		iLog.Log(fpath)

		err = testImage(iWri, fpath, mode)
		if err != nil {
			iLog.Log(err)
		}
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
