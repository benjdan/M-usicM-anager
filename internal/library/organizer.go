package library

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bogem/id3v2/v2"
	"github.com/dhowden/tag"
	"M-usicM-anager/internal/models"
)


type Organizer struct {

	Root string
}


func NewOrganizer(root string) *Organizer {
	return &Organizer{Root: root}
}


func (o *Organizer) OrganizeTrack(srcPath string, track *models.Track, album *models.Album, artist *models.Artist) (string, error) {

	albumFolder := fmt.Sprintf("%s (%s)", album.Title, album.ReleaseDate.Format("2006"))
	destDir := filepath.Join(o.Root, sanitize(artist.Name), sanitize(albumFolder))


	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}


	ext := strings.ToLower(filepath.Ext(srcPath))
	filename := fmt.Sprintf("%02d - %s%s", track.TrackNumber, sanitize(track.Title), ext)
	destPath := filepath.Join(destDir, filename)


	if err := copyFile(srcPath, destPath); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}


	if err := o.writeTags(destPath, track, album, artist); err != nil {
		// Non-fatal — file is moved, tags just didn't write
		fmt.Printf("Warning: failed to write tags for %s: %v\n", destPath, err)
	}


	if err := os.Remove(srcPath); err != nil {
		fmt.Printf("Warning: failed to remove source file %s: %v\n", srcPath, err)
	}

	return destPath, nil
}


func (o *Organizer) writeTags(path string, track *models.Track, album *models.Album, artist *models.Artist) error {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".mp3":
		return o.writeMP3Tags(path, track, album, artist)
	case ".flac":

		return nil
	default:

		return nil
	}
}


func (o *Organizer) writeMP3Tags(path string, track *models.Track, album *models.Album, artist *models.Artist) error {
	t, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open MP3 for tagging: %w", err)
	}
	defer t.Close()

	t.SetTitle(track.Title)
	t.SetArtist(artist.Name)
	t.SetAlbum(album.Title)
	t.SetYear(album.ReleaseDate.Format("2006"))
	t.SetGenre(strings.Join(artist.Genres, ", "))


	t.AddTextFrame(
		t.CommonID("Track number/Position in set"),
		id3v2.EncodingUTF8,
		fmt.Sprintf("%d/%d", track.TrackNumber, album.TotalTracks),
	)


	t.AddTextFrame(
		t.CommonID("Part of a set"),
		id3v2.EncodingUTF8,
		fmt.Sprintf("%d", track.DiscNumber),
	)

	if err := t.Save(); err != nil {
		return fmt.Errorf("failed to save MP3 tags: %w", err)
	}

	return nil
}


func (o *Organizer) ReadTags(path string) (*tag.Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read tags: %w", err)
	}

	return &m, nil
}


func (o *Organizer) ScanLibrary() ([]string, error) {
	var files []string

	err := filepath.Walk(o.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == "downloads" {
			return filepath.SkipDir
		}
		if !info.IsDir() && isAudioFile(path) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}


func sanitize(s string) string {

	re := regexp.MustCompile(`[<>:"/\\|?*]`)
	s = re.ReplaceAllString(s, "")

	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}


func isAudioFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp3", ".flac", ".ogg", ".opus", ".m4a", ".aac", ".wav":
		return true
	}
	return false
}


func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}