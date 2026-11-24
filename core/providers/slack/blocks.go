package slack

import (
	"encoding/json"
	"strconv"
	"strings"
)

// Apps and bots put a message's body into Block Kit blocks or legacy
// attachments; its "text" is then only the notification fallback. These
// turn both into plain text the way the rest of the text is shown.

type wTextObj struct {
	Type string `json:"type"` // plain_text, mrkdwn
	Text string `json:"text"`
}

type wBlock struct {
	Type     string          `json:"type"`
	Text     json.RawMessage `json:"text"` // an object; a string in "markdown" blocks
	Fields   []wTextObj      `json:"fields"`
	Elements []wElem         `json:"elements"`
}

// wElem is an element of a rich_text or context block.
type wElem struct {
	Type        string          `json:"type"`
	Text        json.RawMessage `json:"text"`
	Style       json.RawMessage `json:"style"` // text styles; "bullet"/"ordered" in lists
	URL         string          `json:"url"`
	UserID      string          `json:"user_id"`
	ChannelID   string          `json:"channel_id"`
	UsergroupID string          `json:"usergroup_id"`
	Range       string          `json:"range"`
	Name        string          `json:"name"`
	Unicode     string          `json:"unicode"`
	Fallback    string          `json:"fallback"`
	Value       string          `json:"value"`
	Indent      int             `json:"indent"`
	Offset      int             `json:"offset"`
	Elements    []wElem         `json:"elements"`
}

type wAttField struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

// parseBlocks decodes blocks; broken ones are ignored, not an error.
func parseBlocks(raw json.RawMessage) []wBlock {
	if len(raw) == 0 {
		return nil
	}
	var bs []wBlock
	if json.Unmarshal(raw, &bs) != nil {
		return nil
	}
	return bs
}

// blocksCarryBody tells whether blocks hold more than a user's rich text,
// which is what Slack's own "text" already says.
func blocksCarryBody(bs []wBlock) bool {
	for _, b := range bs {
		if b.Type != "rich_text" {
			return true
		}
	}
	return false
}

// blocksText renders blocks; name resolves user ids.
func blocksText(bs []wBlock, name func(string) string) string {
	var parts []string
	for _, b := range bs {
		switch b.Type {
		case "header", "section":
			var lines []string
			if t := textObj(b.Text, name); t != "" {
				lines = append(lines, t)
			}
			for _, f := range b.Fields {
				if t := objText(f, name); t != "" {
					lines = append(lines, t)
				}
			}
			if len(lines) > 0 {
				parts = append(parts, strings.Join(lines, "\n"))
			}
		case "markdown":
			var s string
			if json.Unmarshal(b.Text, &s) == nil && strings.TrimSpace(s) != "" {
				parts = append(parts, strings.TrimSpace(s))
			}
		case "context":
			var words []string
			for _, e := range b.Elements {
				if e.Type == "mrkdwn" || e.Type == "plain_text" {
					if t := objText(wTextObj{Type: e.Type, Text: rawString(e.Text)}, name); t != "" {
						words = append(words, t)
					}
				}
			}
			if len(words) > 0 {
				parts = append(parts, strings.Join(words, " "))
			}
		case "rich_text":
			if t := strings.TrimSpace(richText(b.Elements, name)); t != "" {
				parts = append(parts, t)
			}
		case "divider":
			if len(parts) > 0 {
				parts = append(parts, "")
			}
		}
		// image, actions, input, file, video: nothing to read.
	}
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return strings.Join(parts, "\n")
}

func rawString(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) != nil {
		return ""
	}
	return s
}

// textObj renders a composition text object given raw.
func textObj(raw json.RawMessage, name func(string) string) string {
	var o wTextObj
	if len(raw) == 0 || json.Unmarshal(raw, &o) != nil {
		return ""
	}
	return objText(o, name)
}

func objText(o wTextObj, name func(string) string) string {
	if o.Type == "mrkdwn" {
		return strings.TrimSpace(plainText(o.Text, name))
	}
	return strings.TrimSpace(o.Text)
}

// richText renders the elements of a rich_text block.
func richText(es []wElem, name func(string) string) string {
	var b strings.Builder
	for _, e := range es {
		switch e.Type {
		case "rich_text_section":
			b.WriteString(richInline(e.Elements, name))
		case "rich_text_preformatted":
			ensureLine(&b)
			b.WriteString("```\n" + strings.TrimRight(richInline(e.Elements, name), "\n") + "\n```\n")
		case "rich_text_quote":
			ensureLine(&b)
			for _, l := range strings.Split(strings.TrimRight(richInline(e.Elements, name), "\n"), "\n") {
				b.WriteString("> " + l + "\n")
			}
		case "rich_text_list":
			ensureLine(&b)
			ordered := rawString(e.Style) == "ordered"
			pad := strings.Repeat("  ", e.Indent)
			for i, item := range e.Elements {
				mark := "• "
				if ordered {
					mark = strconv.Itoa(e.Offset+i+1) + ". "
				}
				b.WriteString(pad + mark + strings.TrimRight(richInline(item.Elements, name), "\n") + "\n")
			}
		}
	}
	return b.String()
}

func ensureLine(b *strings.Builder) {
	if s := b.String(); s != "" && !strings.HasSuffix(s, "\n") {
		b.WriteString("\n")
	}
}

// richInline renders the inline elements of a rich text section in the
// mrkdwn way Slack writes the same message's "text".
func richInline(es []wElem, name func(string) string) string {
	var b strings.Builder
	for _, e := range es {
		switch e.Type {
		case "text":
			b.WriteString(styled(rawString(e.Text), e.Style))
		case "link":
			t := rawString(e.Text)
			if t != "" && t != e.URL {
				b.WriteString(styled(t, e.Style) + " (" + e.URL + ")")
			} else {
				b.WriteString(strings.TrimPrefix(e.URL, "mailto:"))
			}
		case "user":
			b.WriteString("@" + name(e.UserID))
		case "channel":
			b.WriteString("#" + e.ChannelID)
		case "usergroup":
			b.WriteString("@" + e.UsergroupID)
		case "broadcast":
			b.WriteString("@" + e.Range)
		case "emoji":
			b.WriteString(emojiOf(e.Name))
		case "date":
			b.WriteString(e.Fallback)
		case "color":
			b.WriteString(e.Value)
		}
	}
	return b.String()
}

// styled wraps text in the mrkdwn marks of its style.
func styled(t string, raw json.RawMessage) string {
	var st struct{ Bold, Italic, Strike, Code bool }
	if len(raw) == 0 || raw[0] != '{' || json.Unmarshal(raw, &st) != nil || strings.TrimSpace(t) == "" {
		return t
	}
	if st.Code {
		t = "`" + t + "`"
	}
	if st.Strike {
		t = "~" + t + "~"
	}
	if st.Italic {
		t = "_" + t + "_"
	}
	if st.Bold {
		t = "*" + t + "*"
	}
	return t
}

// attachmentText renders a legacy attachment (not a shared message or a
// link unfurl): pretext, author, title, text, "Title: value" fields and
// its blocks, else its fallback.
func attachmentText(a wAttachment, name func(string) string) string {
	var lines []string
	add := func(s string) {
		if s = strings.TrimSpace(plainText(s, name)); s != "" {
			lines = append(lines, s)
		}
	}
	add(a.Pretext)
	add(a.AuthorName)
	if a.Title != "" && a.TitleLink != "" {
		add(a.Title + " (" + a.TitleLink + ")")
	} else {
		add(a.Title)
	}
	add(a.Text)
	for _, f := range a.Fields {
		switch {
		case f.Title != "" && f.Value != "":
			add(f.Title + ": " + f.Value)
		default:
			add(f.Title + f.Value)
		}
	}
	if t := blocksText(parseBlocks(a.Blocks), name); t != "" {
		lines = append(lines, t)
	}
	add(a.Footer)
	if len(lines) == 0 {
		add(a.Fallback)
	}
	return strings.Join(lines, "\n")
}
