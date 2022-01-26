package markdown

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Writer interface {
	For(filename string) io.Writer
}

func NewWriter(path string) Writer {
	if path == "" {
		return &stdio{}
	}
	if err := mkdir(filepath.Join(path)); err != nil {
		exit(err)
	}
	return &writer{
		path: path,
	}
}

type stdio struct{}

func (*stdio) For(filename string) io.Writer {
	return os.Stdout
}

type writer struct {
	path string
	file *os.File
}

func (w *writer) For(filename string) io.Writer {
	var err error
	if w.file != nil {
		if err = w.file.Close(); err != nil {
			panic(err)
		}
	}

	if w.file, err = os.Create(filepath.Join(w.path, filename)); err != nil {
		panic(err)
	}

	return w.file
}

func isFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func mkdir(path string) error {
	b, err := isFileExists(path)
	if err != nil {
		return err
	}

	if !b {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func exit(a ...interface{}) {
	fmt.Println(a...)
	os.Exit(1)
}
