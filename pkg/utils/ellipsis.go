package utils

func Ellipsis(text string, end int) string {
	if end < 0 {
		return ""
	}

	if len(text) <= end {
		return text[:end]
	}

	return text
}
