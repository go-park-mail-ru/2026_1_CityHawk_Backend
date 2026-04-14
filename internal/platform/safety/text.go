package safety

import "html"

func EscapeText(value string) string {
	return html.EscapeString(value)
}

func EscapeTextPtr(value *string) *string {
	if value == nil {
		return nil
	}

	escaped := EscapeText(*value)
	return &escaped
}
