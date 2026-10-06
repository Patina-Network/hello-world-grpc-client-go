package httpserve

import (
	"fmt"
	"mime"
	"strings"
	"time"
)

func isJSONContentType(header string) bool {
	mediaType, _, err := mime.ParseMediaType(header)
	if err != nil {
		return false
	}

	subtype, ok := strings.CutPrefix(mediaType, "application/")

	return ok && (subtype == "json" || strings.HasSuffix(subtype, "+json"))
}

func formatTimestamp(t time.Time) string {
	t = t.UTC()

	var b []byte
	switch year := t.Year(); {
	case year > 9999:
		b = fmt.Appendf(b, "+%d", year)
	case year < 0:
		b = fmt.Appendf(b, "%05d", year)
	default:
		b = fmt.Appendf(b, "%04d", year)
	}

	b = t.AppendFormat(b, "-01-02T15:04:05")

	switch ns := t.Nanosecond(); {
	case ns == 0:
	case ns%1_000_000 == 0:
		b = fmt.Appendf(b, ".%03d", ns/1_000_000)
	case ns%1_000 == 0:
		b = fmt.Appendf(b, ".%06d", ns/1_000)
	default:
		b = fmt.Appendf(b, ".%09d", ns)
	}

	return string(append(b, 'Z'))
}
