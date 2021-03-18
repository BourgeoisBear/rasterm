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
	- iterm2
	- https://www.macterm.net/
  - mintty

## Notes

### terminal features matrix

| terminal | sixel | iTerm2 format | kitty format |
| :---     | :--:  | :--:          | :--:         |
| aminal   | Y     |               |              |
| iterm2   | Y     | Y             |              |
| kitty    |       |               | Y            |
| mintty   | Y     | Y             |              |
| mlterm   | Y     | Y             |              |
| putty    |       |               |              |
| rlogin   | Y     | Y             |              |
| wezterm  | Y     | Y             |              |
| xterm    | Y     |               |              |

### known responses

#### CSI 0 c

| terminal | response                                            |
| :----    | :----                                               |
| kitty    | "\x1b[?62;c"                                        |
| guake    | "\x1b[?65;1;9c"                                     |
| mintty   | "\x1b[?64;1;2;4;6;9;15;21;22;28;29c"                |
| mlterm   | "\x1b[?63;1;2;3;4;7;29c"                            |
| putty    | "\x1b[?6c"                                          |
| rlogin   | "\x1b[?65;1;2;3;4;6;7;8;9;15;18;21;22;29;39;42;44c" |
| st       | "\x1b[?6c"                                          |
| wez      | "\x1b[?65;4;6;18;22c"                               |
| xfce     | "\x1b[?65;1;9c"                                     |
| xterm    | "\x1b[?63;1;2;4;6;9;15;22c"                         |

#### CSI > 0 c

| terminal | response            |
| :----    | :----               |
| kitty    | "\x1b[>1;4000;19c"  |
| guake    | "\x1b[>65;5402;1c"  |
| mintty   | "\x1b[>77;30104;0c" |
| mlterm   | "\x1b[>24;279;0c"   |
| putty    | "\x1b[>0;136;0c"    |
| rlogin   | "\x1b[>65;331;0c"   |
| st       | NO RESPONSE         |
| wez      | "\x1b[>0;0;0c"      |
| xfce     | "\x1b[>65;5402;1c"  |
| xterm    | "\x1b[>19;344;0c    |

### opinions

- Sixel is a primitive and wasteful format.  Most sixel terminals also support the iTerm2 format--fewer bytes, full color instead of paletted, and no pixel re-processing required.  Much better!

### go stuff

```sh
go tool pprof -http=:8080 ./name.prof
godoc -http=:8099 -goroot="$HOME/go"
go test -v
go mod tidy
https://blog.golang.org/pprof
```

### more reading

- kitty inline images:  https://sw.kovidgoyal.net/kitty/graphics-protocol.html
- iterm2 inline images: https://iterm2.com/documentation-images.html
- xterm ctl seqs:       https://invisible-island.net/xterm/ctlseqs/ctlseqs.html
- sixel ctl seqs:       https://vt100.net/docs/vt3xx-gp/chapter14.html
- libsixel:             https://saitoha.github.io/libsixel/
