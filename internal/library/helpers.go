package library

import (
	"strings"
	"time"

	"M-usicM-anager/internal/models"
)

func ParsePartialDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}

	formats := []string{
		"2026-01-02",
		"2026-01",
		"2026",
	}

	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}

	return time.Time{}
}

func NormalizeAlbumType(t string) models.AlbumType {
	switch strings.ToLower(t) {
	case "album":
		return models.AlbumTypeAlbum
	case "ep":
		return models.AlbumTypeEP
	case "single":
		return models.AlbumTypeSingle
	case "live":
		return models.AlbumTypeLive
	case "compilation":
		return models.AlbumTypeCompilation
	default:
		return models.AlbumTypeAlbum
	}
}
