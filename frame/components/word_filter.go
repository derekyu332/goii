package components

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/anknown/ahocorasick"
	"io"
	"os"
	"strings"
)

type WordFilter struct {
	machine  *goahocorasick.Machine
	FileName string
}

func readRunes(filename string) ([][]rune, error) {
	dict := [][]rune{}

	f, err := os.OpenFile(filename, os.O_RDONLY, 0660)

	if err != nil {
		return nil, err
	}

	r := bufio.NewReader(f)

	for {
		l, err := r.ReadBytes('\n')
		if err != nil || err == io.EOF {
			break
		}
		l = bytes.TrimSpace(l)
		dict = append(dict, bytes.Runes(l))
	}

	return dict, nil
}

func (this *WordFilter) Initialize() error {
	dict, err := readRunes(this.FileName)

	if err != nil {
		return err
	}

	m := new(goahocorasick.Machine)

	if err := m.Build(dict); err != nil {
		return err
	}

	this.machine = m

	return nil
}

func (this *WordFilter) IsSensitiveWord(word string) (bool, error) {
	check_word := strings.ToLower(word)
	content := []rune(check_word)
	terms := this.machine.MultiPatternSearch(content, true)

	for _, t := range terms {
		return true, errors.New(string(t.Word))
	}

	return false, nil
}
