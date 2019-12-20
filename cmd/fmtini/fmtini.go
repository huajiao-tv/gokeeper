package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type lineType int

const (
	lineTypeComment = iota
	lineTypeSection
	lineTypeEmpty
	lineTypeKeyValue
	lineTypeLastline
)

const (
	LineBreak  = "\n"
	LineIndent = " "
)

// format fileInput to fileOutput if fileOutput is empty ,overwrite file input
func fmtFile(fileInput, fileOutput string) error {
	if fileInput == "" || fileOutput == "" {
		return fmt.Errorf("invalid infile: %s and outfile: %s", fileInput, fileOutput)
	}
	var inFile, outFile *os.File
	var err error
	outbuf := new(bytes.Buffer)
	inFile, err = os.Open(fileInput) // For read access.
	if err != nil {
		return fmt.Errorf("open %s fail: %s", fileInput, err)
	}
	if err := format(inFile, outbuf); err != nil {
		return fmt.Errorf("format input file: %s error: %s", fileInput, err)
	}
	inFile.Close()
	outFile, err = os.OpenFile(fileOutput, os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("open file: %s error: %s", fileOutput, err)
	}
	defer outFile.Close()
	_, err = outFile.Write(outbuf.Bytes())
	if err != nil {
		return fmt.Errorf("write file %s error: %s", fileInput, err)
	}
	return nil
}

// format a path recursivly
func fmtDir(in string) error {
	f, err := os.Open(in)
	if err != nil {
		return err
	}
	fis, err := f.Readdir(0) // read file info slice fis
	if err != nil {
		return err
	}
	for _, fi := range fis {
		if fi.Name()[0] == '.' {
			fmt.Printf("ignore file or path: %s\n", path.Join(in, fi.Name()))
			continue
		}
		f := path.Join(in, fi.Name())
		if fi.IsDir() {
			if err := fmtDir(f); err != nil {
				return err
			}
			continue
		}
		if err := fmtFile(f, f); err != nil {
			return err
		}
		fmt.Printf("format success: %s\n", f)
	}
	return nil
}

type line struct {
	key              string
	indentToValueTyp int
	valueTyp         string
	indentToValue    int
	value            string
	typ              lineType
}

type iniText struct {
	lines        []line // line of the text
	needToIndent []int  // the line need to be indent
}

func (i *iniText) advance(text string, typ lineType) error {
	switch typ {
	case lineTypeComment:
		i.lines = append(i.lines, line{value: text, typ: typ})
	case lineTypeEmpty, lineTypeSection:
		i.lines = append(i.lines, line{value: text, typ: typ})
		i.adjustLines()
	case lineTypeKeyValue:
		pos := strings.IndexAny(text, "=:")
		if pos < 0 {
			return fmt.Errorf("parsing line: key <type> = value delimiter not found: %s", text)
		} else if pos == 0 {
			return fmt.Errorf("parsing line: key is empty: %s", text)
		}
		l := line{typ: typ, indentToValueTyp: 0, indentToValue: 1}
		left := strings.TrimSpace(text[0:pos])
		fields := strings.Fields(left)
		// l.valueTyp = "string"
		if len(fields) == 2 {
			l.valueTyp = strings.TrimSpace(fields[1])
		}
		if len(fields) == 3 && strings.TrimSpace(fields[2]) == "json" {
			l.valueTyp = strings.TrimSpace(fields[1] + " json")
		}
		l.key = fields[0]
		l.value = strings.TrimSpace(text[pos+1:])
		i.lines = append(i.lines, l)
		i.needToIndent = append(i.needToIndent, len(i.lines)-1)
	case lineTypeLastline:
		if len(i.lines) == 0 {
			break
		}
		var pos int
		for pos = len(i.lines) - 1; pos != 0; pos-- {
			if i.lines[pos].typ != lineTypeEmpty {
				break
			}
		}
		i.lines = i.lines[0 : pos+1]
		i.adjustLines()
	}
	return nil
}

func (i *iniText) adjustLines() {
	if len(i.needToIndent) == 0 {
		return
	}
	maxLength := 0

	for _, index := range i.needToIndent {
		if len(i.lines[index].valueTyp) > maxLength {
			maxLength = len(i.lines[index].valueTyp)
		}
	}
	for _, index := range i.needToIndent {
		i.lines[index].indentToValue += maxLength - len(i.lines[index].valueTyp)
	}
	maxLength = 0
	for _, index := range i.needToIndent {
		if len(i.lines[index].key) > maxLength {
			maxLength = len(i.lines[index].key)
		}
	}
	haveValueType := false
	for _, index := range i.needToIndent {
		if len(i.lines[index].valueTyp) != 0 {
			haveValueType = true
		}
		i.lines[index].indentToValueTyp += maxLength - len(i.lines[index].key)
	}
	if haveValueType {
		for _, index := range i.needToIndent {
			i.lines[index].indentToValueTyp += 1
		}
	}
	i.needToIndent = []int{}
}

func (i iniText) WriteTo(w io.Writer) error {
	buf := bufio.NewWriter(w)
	for _, line := range i.lines {
		switch line.typ {
		case lineTypeComment:
			buf.WriteString(line.value + LineBreak)
		case lineTypeEmpty:
			buf.WriteString(LineBreak)
		case lineTypeSection:
			buf.WriteString(line.value + LineBreak)
		case lineTypeKeyValue:
			buf.WriteString(line.key)
			for i := 0; i < line.indentToValueTyp; i++ {
				buf.WriteString(LineIndent)
			}
			buf.WriteString(line.valueTyp)
			for i := 0; i < line.indentToValue; i++ {
				buf.WriteString(LineIndent)
			}
			buf.WriteString("=" + LineIndent + line.value + LineBreak)
		}
	}
	buf.Flush()
	return nil
}

func (i iniText) String() string {
	b := new(bytes.Buffer)
	i.WriteTo(b)
	return b.String()
}

func format(src io.Reader, dst io.Writer) error {
	textToFormate := iniText{}
	stream := bufio.NewReader(src)
	for {
		l, err := stream.ReadString('\n')
		l = strings.TrimSpace(l)
		length := len(l)
		if err != nil {
			if err != io.EOF {
				return err
			}
			// lastline is empty
			if length == 0 {
				textToFormate.advance(l, lineTypeLastline)
				break
			}
		}
		if length == 0 {
			textToFormate.advance(l, lineTypeEmpty)
			continue
		}
		switch {
		case l[0] == '#' || l[0] == ';': // Comments.
			textToFormate.advance(l, lineTypeComment)
			continue
		case l[0] == '[' && l[length-1] == ']': // New sction.
			textToFormate.advance(l, lineTypeSection)
			continue
		}
		textToFormate.advance(l, lineTypeKeyValue)
	}
	if err := textToFormate.WriteTo(dst); err != nil {
		return err
	}
	return nil
}
