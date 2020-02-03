package controllers

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	Byte = 1 << (iota * 10)
	KiByte
	MiByte
	GiByte
	TiByte
	PiByte
	EiByte
)

// Convert byte size in uint64 to human readable format with IEC binary prefix
func ToHumanReadableBytes(bytes uint64) string {
	unit := ""
	value := float64(bytes)

	switch {
	case bytes >= EiByte:
		unit = "Ei"
		value = value / EiByte
	case bytes >= PiByte:
		unit = "Pi"
		value = value / PiByte
	case bytes >= TiByte:
		unit = "Ti"
		value = value / TiByte
	case bytes >= GiByte:
		unit = "Gi"
		value = value / GiByte
	case bytes >= MiByte:
		unit = "Mi"
		value = value / MiByte
	case bytes >= KiByte:
		unit = "Ki"
		value = value / KiByte
	case bytes >= Byte:
		unit = "B"
	case bytes == 0:
		return "0B"
	}
	result := strconv.FormatFloat(value, 'f', 1, 64)
	result = strings.TrimSuffix(result, ".0")
	return result + unit
}

// Convert human readable byte size format in IEC binary prefix to uint64
func FromHumanReadableBytes(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	i := strings.IndexFunc(s, unicode.IsLetter)
	if i == -1 {
		return 0, fmt.Errorf("unable to parse human readable byte string: %s", s)
	}

	valueString, unitString := s[:i], s[i:]
	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("unable to parse digits in human readable byte string: %s", s)
	}

	switch unitString {
	case "B", "b":
		return uint64(value), nil
	case "Ki", "KiB":
		return uint64(value * KiByte), nil
	case "Mi", "MiB":
		return uint64(value * MiByte), nil
	case "Gi", "GiB":
		return uint64(value * GiByte), nil
	case "Ti", "TiB":
		return uint64(value * TiByte), nil
	case "Pi", "PiB":
		return uint64(value * PiByte), nil
	case "Ei", "EiB":
		return uint64(value * EiByte), nil
	default:
		return 0, fmt.Errorf("unable to parse unit string in human readable byte string: %s", s)
	}
}
