package storage

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewTrackPath(t *testing.T) {
	album := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	re := regexp.MustCompile(`^11111111-2222-3333-4444-555555555555/[0-9a-f-]{36}(\.[a-z0-9]+)?$`)

	tests := []struct {
		clientPath string
		wantExt    string
	}{
		{"x/01-abc.mp3", ".mp3"},
		{"SONG.FLAC", ".flac"},
		{"../../etc/passwd", ""},
		{"../../keys/private.pem", ""},
		{"evil.sh", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := NewTrackPath(album, tt.clientPath)
		if !re.MatchString(got) {
			t.Errorf("NewTrackPath(%q) = %q: unexpected format", tt.clientPath, got)
		}
		if filepath.Ext(got) != tt.wantExt {
			t.Errorf("NewTrackPath(%q) = %q: ext %q, want %q", tt.clientPath, got, filepath.Ext(got), tt.wantExt)
		}
		if !filepath.IsLocal(got) {
			t.Errorf("NewTrackPath(%q) = %q: not local", tt.clientPath, got)
		}
	}

	if NewTrackPath(album, "a.mp3") == NewTrackPath(album, "a.mp3") {
		t.Error("NewTrackPath must generate unique names")
	}
}

func TestNewCoverPath(t *testing.T) {
	album := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	re := regexp.MustCompile(`^11111111-2222-3333-4444-555555555555/[0-9a-f-]{36}(\.[a-z0-9]+)?$`)

	tests := []struct {
		clientPath string
		wantExt    string
	}{
		{"art/cover.jpg", ".jpg"},
		{"COVER.PNG", ".png"},
		{"photo.webp", ".webp"},
		{"../../etc/passwd", ""},
		{"evil.sh", ""},
		{"song.mp3", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := NewCoverPath(album, tt.clientPath)
		if !re.MatchString(got) {
			t.Errorf("NewCoverPath(%q) = %q: unexpected format", tt.clientPath, got)
		}
		if filepath.Ext(got) != tt.wantExt {
			t.Errorf("NewCoverPath(%q) = %q: ext %q, want %q", tt.clientPath, got, filepath.Ext(got), tt.wantExt)
		}
		if !filepath.IsLocal(got) {
			t.Errorf("NewCoverPath(%q) = %q: not local", tt.clientPath, got)
		}
	}

	if NewCoverPath(album, "a.jpg") == NewCoverPath(album, "a.jpg") {
		t.Error("NewCoverPath must generate unique names")
	}
}

func TestAddFileRejectsUnsafePath(t *testing.T) {
	base := t.TempDir()
	s := NewTrackStorage(filepath.Join(base, "tracks"))

	for _, p := range []string{"../escape.mp3", "a/../../escape.mp3", "/abs.mp3", ""} {
		if _, err := s.AddFile(p, strings.NewReader("data")); !errors.Is(err, ErrInvalidPath) {
			t.Errorf("AddFile(%q) err = %v, want ErrInvalidPath", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(base, "escape.mp3")); !os.IsNotExist(err) {
		t.Error("file was written outside of storage")
	}

	n, err := s.AddFile("album/track.mp3", strings.NewReader("data"))
	if err != nil || n != 4 {
		t.Fatalf("AddFile(valid) = %d, %v", n, err)
	}
}

func TestOpen(t *testing.T) {
	base := t.TempDir()
	s := NewTrackStorage(base)
	if _, err := s.AddFile("album/cover.jpg", strings.NewReader("jpeg-bytes")); err != nil {
		t.Fatalf("AddFile: %v", err)
	}

	file, info, err := s.Open("album/cover.jpg")
	if err != nil {
		t.Fatalf("Open(valid) err = %v", err)
	}
	defer file.Close()
	if info.Size() != 10 {
		t.Errorf("Open(valid) size = %d, want 10", info.Size())
	}
	data, err := io.ReadAll(file)
	if err != nil || string(data) != "jpeg-bytes" {
		t.Errorf("Open(valid) content = %q, %v", data, err)
	}

	// Путь за пределами хранилища и каталог вместо файла — ErrInvalidPath,
	// отсутствующий файл — обычная ошибка ФС (обрабатывается как 404).
	for _, p := range []string{"../escape.mp3", "/abs.mp3", "", "album"} {
		if _, _, err := s.Open(p); !errors.Is(err, ErrInvalidPath) {
			t.Errorf("Open(%q) err = %v, want ErrInvalidPath", p, err)
		}
	}
	if _, _, err := s.Open("album/missing.jpg"); err == nil || errors.Is(err, ErrInvalidPath) {
		t.Errorf("Open(missing) err = %v, want fs error", err)
	}
}
