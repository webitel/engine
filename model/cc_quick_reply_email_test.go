package model_test

import (
	"strings"
	"testing"

	"github.com/webitel/engine/model"
)

func appErrId(err model.AppError) string {
	if err == nil {
		return ""
	}
	return err.GetId()
}

func TestQuickReply_PrepareEmail(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, text, html, want, err string
		teams, queues               []int
	}{
		{name: "plain only", text: "Hello ${agent_name}"},
		{name: "formatting kept", text: "x",
			html: `<p>Hi <b>${contact_name}</b>, <a href="https://example.org/a">ref</a></p><ul><li>${thread_id}</li></ul>`,
			want: `<p>Hi <b>${contact_name}</b>, <a href="https://example.org/a">ref</a></p><ul><li>${thread_id}</li></ul>`},
		{name: "active content and styles removed", text: "x",
			html: `<p style="color:red" onclick="x()">A<script>alert(1)</script><img src="https://e/x.png"></p><a href="javascript:alert(1)">B</a>`,
			want: `<p>A</p>B`},
		{name: "markup without text cleared", text: "x", html: "<p><br></p>", want: ""},
		{name: "decoded entity checked", text: "x", html: "<p>&#36;{secret}</p>", err: "quick_reply.email.invalid_token"},
		{name: "unknown variable", text: "${contact_email}", err: "quick_reply.email.invalid_token"},
		{name: "case sensitive", text: "${Agent_name}", err: "quick_reply.email.invalid_token"},
		{name: "unclosed variable", text: "Hi ${agent_name", err: "quick_reply.email.invalid_token"},
		{name: "variable in link", text: "x", html: `<a href="https://e/${thread_id}">t</a>`,
			err: "quick_reply.email.token_in_attribute"},
		{name: "variable in removed attribute ignored", text: "x", html: `<p class="${agent_name}">t</p>`, want: "<p>t</p>"},
		{name: "variable in removed script ignored", text: "x", html: `<p>t</p><script>${x}</script>`, want: "<p>t</p>"},
		{name: "team id", text: "x", teams: []int{0}, err: "quick_reply.email.invalid_restriction"},
		{name: "queue id above integer", text: "x", queues: []int{1 << 31}, err: "quick_reply.email.invalid_restriction"},
		{name: "queue id at integer limit", text: "x", queues: []int{1<<31 - 1}},
		{name: "text required", text: " \n", err: "quick_reply.email.text_required"},
		{name: "text too large", text: strings.Repeat("a", 64<<10+1), err: "quick_reply.email.body_too_large"},
		{name: "html too large", text: "x", html: strings.Repeat("a", 64<<10+1), err: "quick_reply.email.body_too_large"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &model.QuickReply{Channel: "email", Text: tc.text, Html: tc.html, Article: &model.Lookup{Id: 1}}
			for _, id := range tc.teams {
				r.Teams = append(r.Teams, &model.Lookup{Id: id})
			}
			for _, id := range tc.queues {
				r.Queues = append(r.Queues, &model.Lookup{Id: id})
			}
			err := r.PrepareEmail()
			if got := appErrId(err); got != tc.err {
				t.Fatalf("error %q, want %q", got, tc.err)
			}
			if err != nil {
				return
			}
			if r.Html != tc.want {
				t.Fatalf("html %q, want %q", r.Html, tc.want)
			}
			if r.Article != nil {
				t.Fatal("article kept")
			}
		})
	}
}

func TestQuickReply_PatchEmail(t *testing.T) {
	t.Parallel()

	stored := func() *model.QuickReply {
		return &model.QuickReply{Channel: "email", Name: "n", Text: "t", Html: "<p>h</p>",
			Teams: []*model.Lookup{{Id: 1}}, Queues: []*model.Lookup{{Id: 2}}}
	}
	patch := func(fields ...string) *model.QuickReplyPatch {
		return &model.QuickReplyPatch{Fields: fields, Name: model.NewString(""), Text: model.NewString(""),
			Html: model.NewString("")}
	}

	for _, fields := range [][]string{nil, {""}, {"fields", "article.id", "id"}} {
		if _, err := stored().PatchEmail(patch(fields...)); appErrId(err) != "quick_reply.email.empty_patch" {
			t.Fatalf("fields %q: error %q", fields, appErrId(err))
		}
	}

	r := stored()
	p := patch("name", "fields")
	p.Name = model.NewString("renamed")
	fields, err := r.PatchEmail(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 || !fields["name"] {
		t.Fatalf("fields %v", fields)
	}
	if r.Name != "renamed" || r.Text != "t" || r.Html != "<p>h</p>" || len(r.Teams) != 1 || len(r.Queues) != 1 {
		t.Fatalf("name only changed other fields: %+v", r)
	}

	r = stored()
	if _, err := r.PatchEmail(patch("html", "teams")); err != nil {
		t.Fatal(err)
	}
	if r.Html != "" || r.Teams != nil || r.Text != "t" || len(r.Queues) != 1 {
		t.Fatalf("html and teams not cleared: %+v", r)
	}
}

func TestIsValidQuickReplyChannel(t *testing.T) {
	t.Parallel()

	for ch, want := range map[string]bool{"": true, "email": true, "Email": false, "chat": false} {
		if model.IsValidQuickReplyChannel(ch) != want {
			t.Fatalf("channel %q", ch)
		}
	}
}
