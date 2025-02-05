package handler

import (
	"errors"
	"strconv"
	"strings"
)

func ParseRangeHeader(rangeHeader string, fileSize int64) (int64, int64, error) {
	rangeParts := strings.Split(rangeHeader, "=")
	if len(rangeParts) != 2 || rangeParts[0] != "bytes" {
		return 0, 0, errors.New("неверный формат Range-запроса")
	}

	rangeBytes := strings.Split(rangeParts[1], "-")
	start, err := strconv.ParseInt(rangeBytes[0], 10, 64)
	if err != nil {
		return 0, 0, errors.New("ошибка парсинга Range-запроса")
	}

	var end = fileSize - 1
	if len(rangeBytes) > 1 && rangeBytes[1] != "" {
		end, err = strconv.ParseInt(rangeBytes[1], 10, 64)
		if err != nil {
			return 0, 0, errors.New("ошибка парсинга Range-запроса")
		}
	}
	if start > end || end >= fileSize {
		return 0, 0, errors.New("недопустимый диапазон")
	}
	return start, end, nil
}
