package model

import (
	"fmt"
	"math"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
)

const (
	QuickReplyChannelEmail = "email"

	quickReplyEmailMaxBody = 64 << 10
	quickReplyEmailVarOpen = "${"
)

var (
	quickReplyEmailVars = map[string]struct{}{
		"agent_name": {}, "contact_name": {}, "thread_id": {}, "thread_subject": {},
	}
	// quickReplyEmailPatchFields are the PATCH mask paths of an Email reply.
	quickReplyEmailPatchFields = map[string]struct{}{
		"name": {}, "text": {}, "html": {}, "teams": {}, "queues": {},
	}
	quickReplyEmailPolicy = newQuickReplyEmailPolicy()
)

// A small subset of the Email read policy: formatting and static links only.
func newQuickReplyEmailPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "br", "div", "span", "b", "strong", "i", "em", "u", "s",
		"ul", "ol", "li", "blockquote")
	p.AllowAttrs("href").OnElements("a")
	p.AllowURLSchemes("http", "https", "mailto")
	p.RequireParseableURLs(true)
	return p
}

func IsValidQuickReplyChannel(channel string) bool {
	return channel == "" || channel == QuickReplyChannelEmail
}

func NewQuickReplyChannelError(channel string) AppError {
	return NewBadRequestError("quick_reply.email.invalid_channel", fmt.Sprintf("unknown channel %q", channel))
}

func NewQuickReplyRestrictionError() AppError {
	return NewBadRequestError("quick_reply.email.invalid_restriction", "team or queue is not found in the domain")
}

func (r *QuickReply) IsEmail() bool {
	return r.Channel == QuickReplyChannelEmail
}

// PatchEmail applies the supported paths of the mask to an Email reply and
// returns them.
func (r *QuickReply) PatchEmail(patch *QuickReplyPatch) (map[string]bool, AppError) {
	fields := make(map[string]bool, len(patch.Fields))
	for _, f := range patch.Fields {
		if _, ok := quickReplyEmailPatchFields[f]; ok {
			fields[f] = true
		}
	}
	if len(fields) == 0 {
		return nil, NewBadRequestError("quick_reply.email.empty_patch", "no fields to update")
	}

	r.UpdatedAt = patch.UpdatedAt
	r.UpdatedBy = &patch.UpdatedBy
	if fields["name"] && patch.Name != nil {
		r.Name = *patch.Name
	}
	if fields["text"] && patch.Text != nil {
		r.Text = *patch.Text
	}
	if fields["html"] && patch.Html != nil {
		r.Html = *patch.Html
	}
	if fields["teams"] {
		r.Teams = patch.Teams
	}
	if fields["queues"] {
		r.Queues = patch.Queues
	}

	return fields, nil
}

// PrepareEmail validates an Email reply and keeps its sanitized HTML.
func (r *QuickReply) PrepareEmail() AppError {
	r.Article = nil
	if strings.TrimSpace(r.Text) == "" {
		return NewBadRequestError("quick_reply.email.text_required", "text is required")
	}
	if len(r.Text) > quickReplyEmailMaxBody || len(r.Html) > quickReplyEmailMaxBody {
		return quickReplyBodyTooLarge()
	}
	if err := checkQuickReplyVars(r.Text); err != nil {
		return err
	}
	// Queue ids are integer in the database.
	for _, t := range r.Teams {
		if t == nil || t.Id <= 0 {
			return NewQuickReplyRestrictionError()
		}
	}
	for _, q := range r.Queues {
		if q == nil || q.Id <= 0 || q.Id > math.MaxInt32 {
			return NewQuickReplyRestrictionError()
		}
	}

	if r.Html == "" {
		return nil
	}
	// Before sanitizing: the policy drops a link with a variable silently.
	if _, err := checkQuickReplyHTMLVars(r.Html, false); err != nil {
		return err
	}
	clean := quickReplyEmailPolicy.Sanitize(r.Html)
	if len(clean) > quickReplyEmailMaxBody {
		return quickReplyBodyTooLarge()
	}
	visible, err := checkQuickReplyHTMLVars(clean, true)
	if err != nil {
		return err
	}
	if !visible {
		clean = ""
	}
	r.Html = clean

	return nil
}

func quickReplyBodyTooLarge() AppError {
	return NewBadRequestError("quick_reply.email.body_too_large",
		fmt.Sprintf("text and html are limited to %d bytes each", quickReplyEmailMaxBody))
}

// checkQuickReplyVars requires every "${" to start a supported variable.
func checkQuickReplyVars(s string) AppError {
	for {
		i := strings.Index(s, quickReplyEmailVarOpen)
		if i < 0 {
			return nil
		}
		s = s[i+len(quickReplyEmailVarOpen):]
		end := strings.IndexByte(s, '}')
		if end < 0 {
			return NewBadRequestError("quick_reply.email.invalid_token", "unclosed variable")
		}
		if _, ok := quickReplyEmailVars[s[:end]]; !ok {
			return NewBadRequestError("quick_reply.email.invalid_token", fmt.Sprintf("unknown variable %q", s[:end]))
		}
		s = s[end+1:]
	}
}

// checkQuickReplyHTMLVars refuses variables in a[href], or with text in any
// attribute and checks text nodes; it reports whether text is visible.
func checkQuickReplyHTMLVars(s string, text bool) (bool, AppError) {
	visible := false
	z := html.NewTokenizer(strings.NewReader(s))
	for {
		switch z.Next() {
		case html.ErrorToken:
			// Only io.EOF: the input is a string.
			return visible, nil
		case html.TextToken:
			if !text {
				continue
			}
			t := z.Token()
			if strings.TrimSpace(t.Data) != "" {
				visible = true
			}
			if err := checkQuickReplyVars(t.Data); err != nil {
				return false, err
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			for _, a := range t.Attr {
				if !text && (t.Data != "a" || a.Key != "href") {
					continue
				}
				if strings.Contains(a.Val, quickReplyEmailVarOpen) {
					return false, NewBadRequestError("quick_reply.email.token_in_attribute",
						fmt.Sprintf("variable in attribute %q", a.Key))
				}
			}
		}
	}
}
