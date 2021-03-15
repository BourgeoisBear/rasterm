package rasterm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

var (
	ESC_ERASE_DISPLAY = "\x1b[2J\x1b[0;0H"
	E_NON_TTY         = errors.New("NON TTY")
)

// transforms given open/close terminal escapes to pass through tmux to parent terminal
func TmuxOscOpenClose(opn, cls string) (string, string) {

	opn = "\x1bPtmux;" + strings.ReplaceAll(opn, "\x1b", "\x1b\x1b")
	cls = strings.ReplaceAll(cls, "\x1b", "\x1b\x1b") + "\x1b\\"
	return opn, cls
}

func IsTmuxScreen() bool {
	TERM := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	return strings.HasPrefix(TERM, "screen")
}

/*
NOTE: the calling program MUST be connected to an actual terminal for this to work

Requests terminal attributes per:
https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h4-Functions-using-CSI-_-ordered-by-the-final-character-lparen-s-rparen:CSI-Ps-c.1CA3

	CSI Ps c  Send Device Attributes (Primary DA).
		Ps = 0  or omitted ⇒  request attributes from terminal.  The
	response depends on the decTerminalID resource setting.
		⇒  CSI ? 1 ; 2 c     ("VT100 with Advanced Video Option")
		⇒  CSI ? 1 ; 0 c     ("VT101 with No Options")
		⇒  CSI ? 4 ; 6 c     ("VT132 with Advanced Video and Graphics")
		⇒  CSI ? 6 c         ("VT102")
		⇒  CSI ? 7 c         ("VT131")
		⇒  CSI ? 1 2 ; Ps c  ("VT125")
		⇒  CSI ? 6 2 ; Ps c  ("VT220")
		⇒  CSI ? 6 3 ; Ps c  ("VT320")
		⇒  CSI ? 6 4 ; Ps c  ("VT420")

	The VT100-style response parameters do not mean anything by
	themselves.  VT220 (and higher) parameters do, telling the
	host what features the terminal supports:
		Ps = 1    ⇒  132-columns.
		Ps = 2    ⇒  Printer.
		Ps = 3    ⇒  ReGIS graphics.
		Ps = 4    ⇒  Sixel graphics.
		Ps = 6    ⇒  Selective erase.
		Ps = 8    ⇒  User-defined keys.
		Ps = 9    ⇒  National Replacement Character sets.
		Ps = 1 5  ⇒  Technical characters.
		Ps = 1 6  ⇒  Locator port.
		Ps = 1 7  ⇒  Terminal state interrogation.
		Ps = 1 8  ⇒  User windows.
		Ps = 2 1  ⇒  Horizontal scrolling.
		Ps = 2 2  ⇒  ANSI color, e.g., VT525.
		Ps = 2 8  ⇒  Rectangular editing.
		Ps = 2 9  ⇒  ANSI text locator (i.e., DEC Locator mode).
*/
func RequestTermAttributes() (sAttrs []int, E error) {

	// NOTE: raw mode tip came from https://play.golang.org/p/kcMLTiDRZY

	if !term.IsTerminal(syscall.Stdin) {
		return nil, E_NON_TTY
	}

	// STDIN "RAW MODE" TO CAPTURE TERMINAL RESPONSE
	var oldState *term.State
	if oldState, E = term.MakeRaw(syscall.Stdin); E != nil {
		fmt.Println("BOO")
		return
	}
	defer func() {
		// CAPTURE RESTORE ERROR (IF ANY) IF THERE HASN'T ALREADY BEEN AN ERROR
		if e2 := term.Restore(syscall.Stdin, oldState); E != nil {
			E = e2
		}
	}()

	// STDIN NON-BLOCKING MODE IN CASE TERMINAL RESPONSE IS BOGUS
	if E = syscall.SetNonblock(syscall.Stdin, true); E != nil {
		return
	}
	defer syscall.SetNonblock(syscall.Stdin, false)

	// 1/8 SECOND READ DEADLINE TO PREVENT LOCK-UP ON INVALID RESPONSE
	fIN := os.NewFile(uintptr(syscall.Stdin), "stdin")
	if E = fIN.SetReadDeadline(time.Now().Add(time.Second >> 3)); E != nil {
		return
	}

	// SEND REQUEST
	if _, E = os.Stdout.Write([]byte("\x1b[0c")); E != nil {
		return
	}

	// CAPTURE RESPONSE
	reader := bufio.NewReader(fIN)
	text, E := reader.ReadString('c')
	if E != nil {
		return
	}

	// EXTRACT CODES
	pR := regexp.MustCompile(`\d+`)
	t2 := pR.FindAllString(text, -1)
	sAttrs = make([]int, len(t2))
	for ix, sN := range t2 {
		iN, _ := strconv.Atoi(sN)
		sAttrs[ix] = iN
	}

	return
}

func findPtyDevByStat(pStat *syscall.Stat_t) (string, error) {

	for _, devDir := range []string{"/dev/pts", "/dev"} {

		fd, E := os.Open(devDir)
		if os.IsNotExist(E) {
			continue
		} else if E != nil {
			return "", E
		}
		defer fd.Close()

		for {

			fis, e2 := fd.Readdir(256)
			for _, fi := range fis {

				if fi.IsDir() {
					continue
				}

				if s, ok := fi.Sys().(*syscall.Stat_t); ok && (pStat.Dev == s.Dev) && (pStat.Rdev == s.Rdev) && (pStat.Ino == s.Ino) {
					return devDir + "/" + fi.Name(), nil
				}
			}

			if e2 == io.EOF {
				break
			}

			if e2 != nil {
				return "", e2
			}
		}
	}

	return "", os.ErrNotExist
}

func GetTtyPath(pF *os.File) (string, error) {

	info, E := pF.Stat()
	if E != nil {
		return "", E
	}

	if sys, ok := info.Sys().(*syscall.Stat_t); ok {

		fmt.Println("CHECKING: " + pF.Name())
		if path, e := findPtyDevByStat(sys); e == nil {
			return path, nil
		} else if os.IsNotExist(e) {
			return "", E_NON_TTY
		} else {
			return "", e
		}
	}

	return "", nil
}
