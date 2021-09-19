package wad

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// WadType is either PWAD or IWAD, depending on type of WAD indicated in the
// WAD header
type WadType int

const (
	PWAD WadType = iota
	IWAD
)

// Header of the WAD file
type Header struct {
	Type      WadType
	NumFiles  int32
	FatOffset int32
}

// FileEntry in the directory
type FileEntry struct {
	Name   string
	Offset int32
	Length int32
}

// Directory of files in the WAD
type Directory []FileEntry

// GetHeader seeks to the beginning of the given io.ReadSeeker and extracts
// the WAD header, if any.
func GetHeader(r io.ReadSeeker) (*Header, error) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to header: %w", err)
	}

	out := Header{}
	magic := make([]byte, 4)
	_, err := r.Read(magic)
	if err != nil {
		return nil, fmt.Errorf("failed to read magic number: %w", err)
	}

	switch string(magic) {
	case "IWAD":
		out.Type = IWAD
	case "PWAD":
		out.Type = PWAD
	default:
		return nil, fmt.Errorf("unknown magic number '%s'", string(magic))
	}

	if err := binary.Read(r, binary.LittleEndian, &out.NumFiles); err != nil {
		return nil, fmt.Errorf("failed to read NumFiles: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &out.FatOffset); err != nil {
		return nil, fmt.Errorf("failed to read FatOffset: %w", err)
	}

	return &out, nil
}

// Directory seeks to and reads the WAD file info table in the given
// io.ReadSeeker
func (h *Header) Directory(r io.ReadSeeker) (Directory, error) {
	if _, err := r.Seek(int64(h.FatOffset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to directory: %w", err)
	}

	out := make(Directory, 0, h.NumFiles)
	for i := int32(0); i < h.NumFiles; i++ {
		entry, err := getEntry(r)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory entry: %w", err)
		}
		out = append(out, entry)
	}

	return out, nil
}

// Content returns an io.Reader of the lump data coresponding to the receiving
// FileEntry in the given io.ReadSeeker
func (e FileEntry) Content(r io.ReadSeeker) (io.Reader, error) {
	if _, err := r.Seek(int64(e.Offset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to lump data: %w", err)
	}
	return &io.LimitedReader{r, int64(e.Length)}, nil
}

func getEntry(r io.Reader) (FileEntry, error) {
	out := FileEntry{}
	if err := binary.Read(r, binary.LittleEndian, &out.Offset); err != nil {
		return FileEntry{}, fmt.Errorf("failed to read Offset: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &out.Length); err != nil {
		return FileEntry{}, fmt.Errorf("failed to read Length: %w", err)
	}

	nameBytes := make([]byte, 8)
	if _, err := r.Read(nameBytes); err != nil {
		return FileEntry{}, fmt.Errorf("failed to read magic number: %w", err)
	}

	// Truncate nameBytes if zero padded
	if n := bytes.IndexByte(nameBytes, 0); n != -1 {
		nameBytes = nameBytes[:n]
	}
	out.Name = string(nameBytes)

	return out, nil
}
