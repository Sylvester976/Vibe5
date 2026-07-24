package widget

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"spotify-insights/internal/spotify"
)

// signalCoral matches the design tokens in docs/ARCHITECTURE.md §6. The
// widget is deliberately minimal — transparent background, one accent color,
// no card/box — so it reads correctly embedded on either a light GitHub
// README or a dark personal site.
const signalCoral = "#FF6B4A"

// RenderNowPlayingSVG renders a "now playing" badge. np is nil when nothing
// is currently playing.
func RenderNowPlayingSVG(np *spotify.ListeningItem) []byte {
	title := "Not playing right now"
	subtitle := ""
	if np != nil {
		title = np.TrackName
		subtitle = np.ArtistName
	}

	var buf bytes.Buffer
	buf.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="340" height="80" viewBox="0 0 340 80">`)
	fmt.Fprintf(&buf, `<rect x="0" y="34" width="4" height="12" fill="%s"/>`, signalCoral)
	fmt.Fprintf(&buf, `<text x="16" y="34" font-family="sans-serif" font-size="16" font-weight="700" fill="%s">%s</text>`, signalCoral, escapeXML(title))
	if subtitle != "" {
		fmt.Fprintf(&buf, `<text x="16" y="56" font-family="sans-serif" font-size="13" fill="%s">%s</text>`, signalCoral, escapeXML(subtitle))
	}
	buf.WriteString(`</svg>`)
	return buf.Bytes()
}

func escapeXML(s string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(s))
	return buf.String()
}
