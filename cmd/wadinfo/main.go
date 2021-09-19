package main

import (
	"encoding/json"
	"fmt"
	"github.com/boomlinde/wad"
	"os"
)

type info struct {
	Type  string
	Lumps []string
}

func getInfo(fname string) (info, error) {
	f, err := os.Open(fname)
	if err != nil {
		return info{}, fmt.Errorf("failed to open %s: %w", err)
	}
	defer f.Close()

	header, err := wad.GetHeader(f)
	if err != nil {
		return info{}, fmt.Errorf("failed to get header: %w", err)
	}

	out := info{}
	switch header.Type {
	case wad.PWAD:
		out.Type = "PWAD"
	case wad.IWAD:
		out.Type = "IWAD"
	}

	dir, err := header.Directory(f)
	if err != nil {
		return info{}, fmt.Errorf("failed to get directory: %w", err)
	}

	out.Lumps = make([]string, 0, header.NumFiles)
	for _, e := range dir {
		out.Lumps = append(out.Lumps, e.Name)
	}

	return out, nil
}

func main() {
	for _, fname := range os.Args[1:] {
		info, err := getInfo(fname)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(info); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
