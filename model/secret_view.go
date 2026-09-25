package model

import "strings"

// Global secret token(s) view options
var SecretView = secretViewOptions{
	Mask: strings.Repeat("*", 10),
	View: -4, // show last 4 characters
}

type secretViewOptions struct {
	Mask string // suppress with ..
	View int // number of chars to disclose ; + first, - last
}

func (x secretViewOptions) Input(s string) (string, bool) {
	if x.Suppressed(s) {
		return "", false
	}
	return s, true
}

func (x secretViewOptions) Suppress(s string) (vs string) {
	c := len(s)
	if c == 0 {
		// no value
		return ""
	}
	// default: mask
	vs = x.Mask
	// disclose chars ?
	if x.View == 0 {
		// NO ; all are hidden ..
		return vs
	}
	view := x.View
	last := (view < 0)
	if last {
		view *= -1
	}
	view = min(view, c)
	if last {
		vs += s[c-view:]
	} else {
		vs = s[:view] + vs
	}
	return vs
}

func (x secretViewOptions) Suppressed(vs string) (is bool) {
	view := x.View
	last := (view < 0)
	if last {
		view *= -1
		vs, is = strings.CutPrefix(vs, x.Mask)
		return is && len(vs) <= view
	}
	if view > 0 {
		vs, is = strings.CutSuffix(vs, x.Mask)
		return is && len(vs) <= view
	}
	return len(x.Mask) > 0 && vs == x.Mask
}