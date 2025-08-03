package process

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"golang.org/x/sync/errgroup"
)

type processer struct {
	RootDir string
	Dirs    []string
}

func NewProcesser(rootDir string) *processer {
	return &processer{
		RootDir: rootDir,
	}
}

func Processing(p *processer) {
	p.chdirPrep()
	if err := p.getDirs(); err != nil {
		slog.Error("change dir in processing",
			slog.String("err", err.Error()),
		)
	}

	if err := p.processFirstLevel(); err != nil {
		slog.Error("something went wrong in processing first level",
			slog.String("err", err.Error()),
		)
	}

	if err := p.groupByExtension(); err != nil {
		slog.Error("something went wrong when grouping files",
			slog.String("err", err.Error()),
		)
	}
}

func (p *processer) groupByExtension() error {
	entries, err := os.ReadDir(p.RootDir)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}
		ext, err := getExtension(fileInfo.Name())
		if err != nil {
			continue
		}
		if strings.Contains(p.RootDir, ext) {
			continue
		}
		if !slices.Contains(p.Dirs, ext) {
			makeDir(filepath.Join(p.RootDir, ext))
		}
		moveFileTo(
			filepath.Join(p.RootDir, fileInfo.Name()),
			filepath.Join(p.RootDir, ext, fileInfo.Name()),
		)
	}

	return nil
}

func moveFileTo(oldFilePath, newFilePath string) error {
	return os.Rename(oldFilePath, newFilePath)
}

func makeDir(dirName string) error {
	return os.Mkdir(dirName, 0755)
}

func getExtension(ext string) (string, error) {
	extRaw := filepath.Ext(ext)
	if utf8.RuneCountInString(extRaw) == 0 {
		return "", fmt.Errorf("no extension found")
	}
	prepExt := strings.ToLower(strings.ReplaceAll(extRaw, ".", ""))
	return prepExt, nil
}

func (p *processer) processFirstLevel() error {
	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(10)

	for i, dir := range p.Dirs {
		dir := dir
		i := i

		g.Go(func() error {
			localProc := NewProcesser(filepath.Join(p.RootDir, dir))
			if err := localProc.groupByExtension(); err != nil {
				return err
			}
			slog.Info("processing",
				slog.Int("fileIndex", i),
				slog.String("filePath", filepath.Join(p.RootDir, dir)),
			)
			return nil
		})
	}

	return g.Wait()
}

// Change current dir to root dir
func (p *processer) chdirPrep() {
	slog.Info("chdirPrep", slog.String("workDir", p.RootDir))
	err := os.Chdir(p.RootDir)
	if err != nil {
		slog.Error("Error chdirPrep", slog.String("err", err.Error()))
		return
	}
}

// Get all directoris from currect RootDir
func (p *processer) getDirs() error {
	entries, err := os.ReadDir(p.RootDir)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Type().IsDir() {
			p.Dirs = append(p.Dirs, entry.Name())
		}
	}

	return nil
}
