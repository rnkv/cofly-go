package cofly

import (
	"strconv"
	"strings"
)

type span struct {
	indexFrom int
	indexTo   int
}

func newSpan(indexFrom, indexTo int) span {
	return span{
		indexFrom: indexFrom,
		indexTo:   indexTo,
	}
}

func parseSpan(key string) (span, bool) {
	indexFromString, indexToString, found := strings.Cut(key, "..")
	if !found {
		return span{}, false
	}

	indexFrom, err := strconv.ParseInt(indexFromString, 10, 0)
	if err != nil {
		return span{}, false
	}

	indexTo := indexFrom

	if indexToString != "" {
		indexTo, err = strconv.ParseInt(indexToString, 10, 0)
		if err != nil {
			return span{}, false
		}
	}

	if indexFrom > indexTo {
		return span{}, false
	}

	return span{
		indexFrom: int(indexFrom),
		indexTo:   int(indexTo),
	}, true
}

func (s span) length() int {
	return s.indexTo - s.indexFrom
}

func (s span) positiveBeforeLength(length int) (span, bool) {
	if s.indexTo < 0 || s.indexFrom >= length {
		return span{}, false
	}

	return span{
		indexFrom: max(s.indexFrom, 0),
		indexTo:   min(s.indexTo, length),
	}, true
}

func (s span) negative() (span, bool) {
	if s.indexFrom >= 0 {
		return span{}, false
	}

	return span{
		indexFrom: s.indexFrom,
		indexTo:   min(s.indexTo, 0),
	}, true
}

func (s span) afterLength(length int) (span, bool) {
	if s.indexTo < length {
		return span{}, false
	}

	return span{
		indexFrom: max(s.indexFrom, length),
		indexTo:   s.indexTo,
	}, true
}

func (s span) string() string {
	buffer := make([]byte, 0, 24)
	buffer = strconv.AppendInt(buffer, int64(s.indexFrom), 10)
	buffer = append(buffer, '.', '.')

	if s.indexTo > s.indexFrom {
		buffer = strconv.AppendInt(buffer, int64(s.indexTo), 10)
	}

	return string(buffer)
}
