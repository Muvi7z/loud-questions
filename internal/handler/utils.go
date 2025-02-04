package handler

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseRangeHeader(rangeHeader string, fileSize int64) (int64, int64, error) {
	bytesRange := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")

	start, err := strconv.ParseInt(bytesRange[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	var end int64
	if len(bytesRange) > 1 {
		end, err = strconv.ParseInt(bytesRange[1], 10, 64)
		if err != nil {
			return 0, 0, err
		}
	} else {
		end = fileSize - 1
	}

	if start > end || end > fileSize {
		return 0, 0, fmt.Errorf("invalid range")
	}

	return start, end, nil
}
