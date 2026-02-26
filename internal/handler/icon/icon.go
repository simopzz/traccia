// Package icon provides Lucide SVG icon components for use in templ templates.
package icon

import (
	"context"
	"fmt"
	"html"
	"io"

	"github.com/a-h/templ"
)

// Props configures icon rendering.
type Props struct {
	Class string
	Size  int
}

func resolve(props []Props) Props {
	p := Props{Size: 24}
	if len(props) > 0 {
		if props[0].Size > 0 {
			p.Size = props[0].Size
		}
		p.Class = props[0].Class
	}
	return p
}

type svgComponent struct{ html string }

func (s svgComponent) Render(_ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, s.html)
	return err
}

func newIcon(paths string, props []Props) templ.Component {
	p := resolve(props)
	classAttr := ""
	if p.Class != "" {
		classAttr = fmt.Sprintf(` class="%q"`, html.EscapeString(p.Class))
	}
	return svgComponent{html: fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"%s>%s</svg>`,
		p.Size, p.Size, classAttr, paths,
	)}
}

func MapPin(props ...Props) templ.Component {
	return newIcon(`<path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3"/>`, props)
}

func Utensils(props ...Props) templ.Component {
	return newIcon(`<path d="M3 2v7c0 1.1.9 2 2 2h4a2 2 0 0 0 2-2V2"/><path d="M7 2v20"/><path d="M21 15V2v0a5 5 0 0 0-5 5v6c0 1.1.9 2 2 2h3Zm0 0v7"/>`, props)
}

func Bed(props ...Props) templ.Component {
	return newIcon(`<path d="M2 4v16"/><path d="M2 8h18a2 2 0 0 1 2 2v10"/><path d="M2 17h20"/><path d="M6 8v9"/>`, props)
}

func Bus(props ...Props) templ.Component {
	return newIcon(`<path d="M8 6v6"/><path d="M15 6v6"/><path d="M2 12h19.6"/><path d="M18 18h3s.5-1.7.8-2.8c.1-.4.2-.8.2-1.2 0-.4-.1-.8-.2-1.2l-1.4-5C20.1 6.8 19.1 6 18 6H4a2 2 0 0 0-2 2v10h3"/><circle cx="7" cy="18" r="2"/><path d="M9 18h5"/><circle cx="16" cy="18" r="2"/>`, props)
}

func Plane(props ...Props) templ.Component {
	return newIcon(`<path d="M2 22h20"/><path d="M6.36 17.4 4 17l-2-4 1.1-.55a2 2 0 0 1 1.8 0l.17.1a2 2 0 0 0 1.8 0L8 12 5 6l.9-.45a2 2 0 0 1 2.09.2l4.02 3a2 2 0 0 0 2.1.2l4.19-2.06a2.41 2.41 0 0 1 1.73-.17L21 7a1.4 1.4 0 0 1 .87 1.99l-.38.76c-.23.46-.6.84-1.07 1.08L7.58 17.2a2 2 0 0 1-1.22.18Z"/>`, props)
}

func Lock(props ...Props) templ.Component {
	return newIcon(`<rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>`, props)
}

func GripVertical(props ...Props) templ.Component {
	return newIcon(`<circle cx="9" cy="12" r="1"/><circle cx="9" cy="5" r="1"/><circle cx="9" cy="19" r="1"/><circle cx="15" cy="12" r="1"/><circle cx="15" cy="5" r="1"/><circle cx="15" cy="19" r="1"/>`, props)
}

func ChevronRight(props ...Props) templ.Component {
	return newIcon(`<path d="m9 18 6-6-6-6"/>`, props)
}
