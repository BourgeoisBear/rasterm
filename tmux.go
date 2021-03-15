package rasterm

import (
	"os"
	"strings"
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
