package telegram

import (
	"sort"

	"github.com/gotd/td/tg"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/richtext"
)

// spansOf maps Telegram's own styles. Auto-detected entities (bare links,
// @names, hashtags) carry no style: the app finds links itself.
func spansOf(ents []tg.MessageEntityClass) []model.TextSpan {
	var out []model.TextSpan
	add := func(k model.SpanKind, e tg.MessageEntityClass, url, lang string) {
		out = append(out, model.TextSpan{Kind: k, Offset: e.GetOffset(), Length: e.GetLength(), URL: url, Lang: lang})
	}
	for _, e := range ents {
		switch x := e.(type) {
		case *tg.MessageEntityBold:
			add(model.SpanBold, e, "", "")
		case *tg.MessageEntityItalic:
			add(model.SpanItalic, e, "", "")
		case *tg.MessageEntityUnderline:
			add(model.SpanUnderline, e, "", "")
		case *tg.MessageEntityStrike:
			add(model.SpanStrike, e, "", "")
		case *tg.MessageEntitySpoiler:
			add(model.SpanSpoiler, e, "", "")
		case *tg.MessageEntityCode:
			add(model.SpanCode, e, "", "")
		case *tg.MessageEntityPre:
			add(model.SpanPre, e, "", x.Language)
		case *tg.MessageEntityBlockquote:
			add(model.SpanQuote, e, "", "")
		case *tg.MessageEntityTextURL:
			if richtext.SafeURL(x.URL) {
				add(model.SpanLink, e, x.URL, "")
			}
		}
	}
	return out
}

// incomingText is a message's text and styles. Telegram's own styles win;
// a text without any may still be written in Markdown.
func incomingText(text string, ents []tg.MessageEntityClass) (string, []model.TextSpan) {
	if spans := spansOf(ents); len(spans) > 0 {
		return text, spans
	}
	d := richtext.Parse(text, richtext.Markdown)
	return d.Text, d.Spans
}

// entitiesOf maps styles to Telegram's entities.
func entitiesOf(spans []model.TextSpan) []tg.MessageEntityClass {
	var out []tg.MessageEntityClass
	for _, s := range spans {
		o, l := s.Offset, s.Length
		switch s.Kind {
		case model.SpanBold, model.SpanHeading: // no headings in Telegram
			out = append(out, &tg.MessageEntityBold{Offset: o, Length: l})
		case model.SpanItalic:
			out = append(out, &tg.MessageEntityItalic{Offset: o, Length: l})
		case model.SpanUnderline:
			out = append(out, &tg.MessageEntityUnderline{Offset: o, Length: l})
		case model.SpanStrike:
			out = append(out, &tg.MessageEntityStrike{Offset: o, Length: l})
		case model.SpanSpoiler:
			out = append(out, &tg.MessageEntitySpoiler{Offset: o, Length: l})
		case model.SpanCode:
			out = append(out, &tg.MessageEntityCode{Offset: o, Length: l})
		case model.SpanPre:
			out = append(out, &tg.MessageEntityPre{Offset: o, Length: l, Language: s.Lang})
		case model.SpanQuote:
			out = append(out, &tg.MessageEntityBlockquote{Offset: o, Length: l})
		case model.SpanLink:
			out = append(out, &tg.MessageEntityTextURL{Offset: o, Length: l, URL: s.URL})
		}
	}
	return out
}

// outgoingText reads src (o.Text, maybe after a "> " quote) as Markdown and
// returns it with its entities, the mentions of o.Text among them.
func (s *session) outgoingText(o model.Outgoing, src string) (richtext.Doc, []tg.MessageEntityClass) {
	shift := len(src) - len(o.Text)
	keep := make([]richtext.Range, 0, len(o.Mentions))
	if shift >= 0 && src[shift:] == o.Text {
		for _, m := range o.Mentions {
			keep = append(keep, richtext.Range{Offset: m.Offset + shift, Length: m.Length})
		}
	}
	d := richtext.Parse(src, richtext.Markdown, keep...)
	var moved []model.Mention
	for i, k := range d.Keep {
		if k.Offset >= 0 {
			m := o.Mentions[i]
			m.Offset = k.Offset
			moved = append(moved, m)
		}
	}
	ents := append(entitiesOf(d.Spans), s.mentionEntities(d.Text, moved)...)
	sort.SliceStable(ents, func(i, j int) bool { return ents[i].GetOffset() < ents[j].GetOffset() })
	return d, ents
}
