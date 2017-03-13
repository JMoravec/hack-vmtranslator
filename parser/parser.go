package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	// CArithmetic is the type returned for all arithmetic commands
	CArithmetic = iota
	// CPush is the type returned for push commands
	CPush = iota
	// CPop is the type returned for pop commands
	CPop = iota
)

// Parser struct that holds the data of the source file
type Parser struct {
	filepath           string
	fileScanner        *bufio.Scanner
	CurrentCommandType int
	currentLine        string
	Arg1               string
	Arg2               int
	splitLine          []string
	file               *os.File
}

// NewParser creates a new parser with the given filepath
func NewParser(filepath string) *Parser {
	newParse := &Parser{filepath: filepath, CurrentCommandType: -1}
	return newParse
}

// OpenFile opens the parser and sets the file scanner
func (p *Parser) OpenFile() error {
	if p.filepath == "" {
		return errors.New("Filepath not set")
	}

	if p.filepath[len(p.filepath)-3:] != ".vm" {
		return errors.New("File must have a .vm ending")
	}
	var err error
	p.file, err = os.Open(p.filepath)
	if err != nil {
		return err
	}
	p.fileScanner = bufio.NewScanner(p.file)
	return nil
}

// Advance moves the parser to the next command
func (p *Parser) Advance() error {
	if p.fileScanner == nil {
		return errors.New("File has not been opened")
	}

	for {
		if !p.fileScanner.Scan() && p.fileScanner.Err() == nil {
			return errors.New("EOF")
		}

		if err := p.fileScanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading file: ", err)
			return err
		}

		if !isCommentOrEmptyLine(p.fileScanner.Text()) {
			p.currentLine = p.fileScanner.Text()
			// reset internals
			p.CurrentCommandType = -1
			p.Arg1 = ""
			p.Arg2 = -1
			p.splitLine = nil
			return nil
		}
		// don't return if the current line is a comment
	}
}

// isComment checks the given string to see if the line of text starts with // or is just whitespace
func isCommentOrEmptyLine(text string) bool {
	whiteSpace := strings.Replace(text, " ", "", -1)
	whiteSpace = strings.Replace(whiteSpace, "\t", "", -1)
	return whiteSpace == "" || (strings.Contains(text, "//") && whiteSpace[0:2] == "//")
}

// SetCommandType sets the command type of the parser depending on the command of the currentLine
func (p *Parser) SetCommandType() error {
	if p.currentLine == "" {
		return errors.New("Current line hasn't been set")
	}

	//remove excess whitespace
	// matches leading/trailing whitespace
	re := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	// matches whitespace inbetween words
	reInside := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	final := re.ReplaceAllString(p.currentLine, "")
	final = reInside.ReplaceAllString(final, " ")
	p.splitLine = strings.Split(final, " ")

	switch strings.ToLower(p.splitLine[0]) {
	case "add":
		fallthrough
	case "sub":
		fallthrough
	case "neg":
		fallthrough
	case "eq":
		fallthrough
	case "gt":
		fallthrough
	case "lt":
		fallthrough
	case "and":
		fallthrough
	case "or":
		fallthrough
	case "not":
		p.CurrentCommandType = CArithmetic
	case "push":
		p.CurrentCommandType = CPush
	case "pop":
		p.CurrentCommandType = CPop
	}

	return nil
}

// SetArg1 sets the parser's Arg1 field to the first argument of the current command line
func (p *Parser) SetArg1() error {
	if p.CurrentCommandType == -1 {
		// set the command type if it hasn't already, but exit if there's an error
		if err := p.SetCommandType(); err != nil {
			return err
		}
	}

	switch p.CurrentCommandType {
	case CArithmetic:
		// CArithmetic type will just use the arthmetic command (add, sub, etc)
		p.Arg1 = strings.ToLower(p.splitLine[0])
	case CPush:
		// CPush will return the memory segement
		fallthrough
	case CPop:
		// CPop will use the memory segment
		p.Arg1 = strings.ToLower(p.splitLine[1])
	}

	return nil
}

// SetArg2 Sets the second argument for the parser. Will only be set if the command type supports it
func (p *Parser) SetArg2() error {
	if p.CurrentCommandType == -1 {
		// set the command type if it hasn't already, but exit if there's an error
		if err := p.SetCommandType(); err != nil {
			return err
		}
	}

	if p.CurrentCommandType == CArithmetic {
		return errors.New("Current command is CArithmetic, no Arg2 exists")
	}

	Arg2, err := strconv.ParseInt(p.splitLine[2], 0, 0)
	if err != nil {
		return err
	}

	p.Arg2 = int(Arg2)
	return nil
}

// Close the file
func (p *Parser) Close() {
	p.file.Close()
}
