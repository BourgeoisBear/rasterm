# rasterm

Encodes images to iTerm / Kitty / SIXEL (terminal) inline graphics protocols.

[![GoDoc](https://godoc.org/github.com/BourgeoisBear/rasterm?status.png)](http://godoc.org/github.com/BourgeoisBear/rasterm)

![rasterm sample output](screenshot.png)

## Supported Image Encodings

- **Kitty**
- **iTerm2 / WezTerm**
- **Sixel**

## TODO

- generic version of RequestTermAttributes()
- mintty:
	- detection for iTerm format: https://github.com/mintty/mintty/issues/881
- iTerm2:
	- support: name, width, height, preserveAspectRatio options
- Kitty:
	- support animation
	- support image placement /dims
- perhaps query tmux directly: TMUX=/tmp/tmux-1000/default,3218,4
- improve terminal identification
	19:VT340
	ESC[>0c = 19;344:0c
	https://invisible-island.net/xterm/ctlseqs/ctlseqs-contents.html

## TESTING

- test sixel with
	- https://github.com/liamg/aminal
	- https://domterm.org/
	- https://www.macterm.net/
- test wez/iterm img with
	- https://www.macterm.net/
  - mintty

## Notes

### go stuff

```sh
go tool pprof -http=:8080 ./name.prof
godoc -http=:8099 -goroot="$HOME/go"
go test -v
go mod tidy
```

### more reading

- kitty inline images:  https://sw.kovidgoyal.net/kitty/graphics-protocol.html
- iterm2 inline images: https://iterm2.com/documentation-images.html
- xterm ctl seqs:       https://invisible-island.net/xterm/ctlseqs/ctlseqs.html
- sixel ctl seqs:       https://vt100.net/docs/vt3xx-gp/chapter14.html
- libsixel:             https://saitoha.github.io/libsixel/
