# rasterm
Encodes images to iTerm / Kitty / SIXEL (terminal) inline graphics protocols.

### TODO
- improve terminal identification
- check that mintty supports iterm/wezterm format, get mintty identifier
- unit tests

### TESTING
- test sixel with
	- https://github.com/liamg/aminal
	- https://domterm.org/
	- https://www.macterm.net/
- test wez/iterm img with
	- https://www.macterm.net/
  - mintty

### Supported Image Encodings
- *WezTerm & iTerm2*: https://iterm2.com/documentation-images.html
- *Kitty*: https://sw.kovidgoyal.net/kitty/graphics-protocol.html
- *Sixel*: https://saitoha.github.io/libsixel/
