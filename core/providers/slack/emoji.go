package slack

import "strings"

// emojiNames maps Slack's names of common emoji to the emoji. Others
// (custom ones included) are shown as ":name:".
var emojiNames = map[string]string{
	"+1": "👍", "thumbsup": "👍", "-1": "👎", "thumbsdown": "👎", "heart": "❤️", "joy": "😂",
	"smile": "😄", "smiley": "😃", "grinning": "😀", "laughing": "😆", "sweat_smile": "😅", "rolling_on_the_floor_laughing": "🤣",
	"slightly_smiling_face": "🙂", "upside_down_face": "🙃", "wink": "😉", "blush": "😊", "innocent": "😇", "heart_eyes": "😍",
	"star-struck": "🤩", "kissing_heart": "😘", "yum": "😋", "stuck_out_tongue": "😛", "stuck_out_tongue_winking_eye": "😜",
	"thinking_face": "🤔", "face_with_raised_eyebrow": "🤨", "neutral_face": "😐", "expressionless": "😑", "no_mouth": "😶",
	"smirk": "😏", "unamused": "😒", "face_with_rolling_eyes": "🙄", "grimacing": "😬", "relieved": "😌", "pensive": "😔",
	"sleepy": "😪", "sleeping": "😴", "mask": "😷", "nauseated_face": "🤢", "exploding_head": "🤯", "sunglasses": "😎",
	"nerd_face": "🤓", "confused": "😕", "worried": "😟", "slightly_frowning_face": "🙁", "open_mouth": "😮", "hushed": "😯",
	"astonished": "😲", "flushed": "😳", "pleading_face": "🥺", "cry": "😢", "sob": "😭", "scream": "😱", "confounded": "😖",
	"disappointed": "😞", "sweat": "😓", "weary": "😩", "tired_face": "😫", "triumph": "😤", "rage": "😡", "angry": "😠",
	"skull": "💀", "poop": "💩", "hankey": "💩", "clown_face": "🤡", "ghost": "👻", "alien": "👽", "robot_face": "🤖",
	"see_no_evil": "🙈", "hear_no_evil": "🙉", "speak_no_evil": "🙊", "wave": "👋", "ok_hand": "👌", "v": "✌️",
	"crossed_fingers": "🤞", "point_up": "☝️", "point_down": "👇", "point_left": "👈", "point_right": "👉", "clap": "👏",
	"raised_hands": "🙌", "open_hands": "👐", "pray": "🙏", "handshake": "🤝", "muscle": "💪", "eyes": "👀", "eye": "👁️",
	"brain": "🧠", "fire": "🔥", "100": "💯", "sparkles": "✨", "star": "⭐", "tada": "🎉", "confetti_ball": "🎊",
	"rocket": "🚀", "white_check_mark": "✅", "heavy_check_mark": "✔️", "ballot_box_with_check": "☑️", "x": "❌",
	"heavy_multiplication_x": "✖️", "warning": "⚠️", "exclamation": "❗", "question": "❓", "bangbang": "‼️", "no_entry": "⛔",
	"no_entry_sign": "🚫", "red_circle": "🔴", "large_green_circle": "🟢", "large_yellow_circle": "🟡", "large_blue_circle": "🔵",
	"zap": "⚡", "boom": "💥", "bulb": "💡", "memo": "📝", "pushpin": "📌", "link": "🔗", "lock": "🔒", "key": "🔑",
	"hourglass": "⌛", "hourglass_flowing_sand": "⏳", "stopwatch": "⏱️", "alarm_clock": "⏰", "calendar": "📆",
	"coffee": "☕", "beer": "🍺", "beers": "🍻", "pizza": "🍕", "cake": "🍰", "birthday": "🎂", "gift": "🎁",
	"trophy": "🏆", "medal": "🏅", "dart": "🎯", "bug": "🐛", "ship": "🚢", "construction": "🚧", "rotating_light": "🚨",
	"chart_with_upwards_trend": "📈", "chart_with_downwards_trend": "📉", "moneybag": "💰", "heavy_plus_sign": "➕",
	"heavy_minus_sign": "➖", "arrow_up": "⬆️", "arrow_down": "⬇️", "arrow_right": "➡️", "arrow_left": "⬅️",
	"repeat": "🔁", "recycle": "♻️", "speech_balloon": "💬", "thought_balloon": "💭", "mag": "🔍", "bell": "🔔",
	"broken_heart": "💔", "orange_heart": "🧡", "yellow_heart": "💛", "green_heart": "💚", "blue_heart": "💙",
	"purple_heart": "💜", "black_heart": "🖤", "white_heart": "🤍", "heart_on_fire": "❤️‍🔥", "hugging_face": "🤗",
	"shushing_face": "🤫", "zipper_mouth_face": "🤐", "partying_face": "🥳", "saluting_face": "🫡", "melting_face": "🫠",
	"face_palm": "🤦", "facepalm": "🤦", "shrug": "🤷", "man-shrugging": "🤷‍♂️", "woman-shrugging": "🤷‍♀️",
	"ok": "🆗", "new": "🆕", "cool": "🆒", "sos": "🆘", "white_circle": "⚪", "black_circle": "⚫", "unicorn_face": "🦄",
	"dog": "🐶", "cat": "🐱", "turtle": "🐢", "snail": "🐌", "sun_with_face": "🌞", "rainbow": "🌈", "snowflake": "❄️",
}

var emojiByChar = func() map[string]string {
	m := make(map[string]string, len(emojiNames))
	for name, e := range emojiNames {
		if old, ok := m[e]; !ok || len(name) < len(old) {
			m[e] = name
		}
	}
	m["👍"], m["👎"] = "+1", "-1"
	return m
}()

// emojiOf turns a reaction name ("+1", "eyes::skin-tone-2") into the emoji.
func emojiOf(name string) string {
	base, _, _ := strings.Cut(name, "::")
	if e, ok := emojiNames[base]; ok {
		return e
	}
	return ":" + name + ":"
}

// emojiName is emojiOf the other way round.
func emojiName(emoji string) string {
	if n, ok := emojiByChar[emoji]; ok {
		return n
	}
	return strings.Trim(emoji, ":")
}
