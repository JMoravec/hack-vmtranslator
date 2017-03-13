package codeWriter

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"path"

	"github.com/jmoravec/VMtranslator/parser"
)

// CodeWriter struct that holds the file information for the translated file
type CodeWriter struct {
	filepath string
	file     *os.File
	jmpSpot  int
}

// NewCodeWriter creates a new CodeWriter object
func NewCodeWriter(filepath string) (*CodeWriter, error) {
	f, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}
	return &CodeWriter{filepath: filepath, file: f}, nil
}

// Close syncs the file and closes it
func (c *CodeWriter) Close() error {
	endCommand := `// end loop command
(END_LOOP)
	@END_LOOP
	0;JMP
`
	_, err := c.file.WriteString(endCommand)
	defer c.file.Close()
	if err == nil {
		err = c.file.Sync()
	}
	return err
}

// WriteArithmetic writes the given arthimetic command (add, sub, neg, etc) to the file
func (c *CodeWriter) WriteArithmetic(command string) error {
	if command == "not" || command == "neg" {
		code := `// ` + command + `
	@SP
	A=M-1
`
		if command == "not" {
			code += "M=!M\n"
		} else {
			code += "M=-M\n"
		}
		if _, err := c.file.WriteString(code); err != nil {
			return err
		}

		return nil
	}
	contents := "// " + command
	contents += `
	@SP
	A=M-1
	D=M
	A=A-1
	`
	var magicCommand string
	switch command {
	case "add":
		magicCommand = "M=M+D"
	case "sub":
		magicCommand = "M=M-D"
	case "neg":
		magicCommand = "M=-D"
	case "eq":
		fallthrough
	case "lt":
		fallthrough
	case "gt":
		jmpSpotStr := strconv.Itoa(c.jmpSpot)
		magicCommand = getEqualityCode(jmpSpotStr, command)
		c.jmpSpot++
	case "and":
		magicCommand = "M=D&M"
	case "or":
		magicCommand = "M=D|M"
	}

	contents += magicCommand
	contents += `
	@SP
	M=M-1
`

	if _, err := c.file.WriteString(contents); err != nil {
		return err
	}

	return nil
}

func getEqualityCode(jmpSpotStr, eqType string) string {
	command := `D=M-D
	@TRUE.` + jmpSpotStr
	switch eqType {
	case "eq":
		command += `
	D;JEQ`
	case "lt":
		command += `
	D;JLT`
	case "gt":
		command += `
	D;JGT`
	}

	command += `
	@FALSE.` + jmpSpotStr + `
	0;JMP
(TRUE.` + jmpSpotStr + `)
	@SP
	A=M-1
	A=A-1
	M=-1
	@CONTINUE.` + jmpSpotStr + `
	0;JMP
(FALSE.` + jmpSpotStr + `)
	@SP
	A=M-1
	A=A-1
	M=0
	@CONTINUE.` + jmpSpotStr + `
	0;JMP
(CONTINUE.` + jmpSpotStr + `) `

	return command
}

// WritePushPop writes the corresponding push or pop command to the file
func (c *CodeWriter) WritePushPop(command int, segment string, index int) error {
	var contents string
	indexStr := strconv.Itoa(index)

	// will get the hack label for the local, this, that, and argument commands
	segLoc, locationCommand := getHackLabel(segment)
	if command == parser.CPush {
		contents += writePush(segment, indexStr, segLoc, locationCommand, c)
	} else if command == parser.CPop {
		contents += writePop(segment, indexStr, segLoc, locationCommand, c)
	} else {
		return errors.New("Push or Pop command not sent")
	}

	if _, err := c.file.WriteString(contents); err != nil {
		return err
	}
	return nil
}

func getHackLabel(vmLabel string) (string, string) {
	hackLabel := ""
	command := ""
	switch vmLabel {
	case "local":
		hackLabel = "LCL"
		command = "D+M"
	case "argument":
		hackLabel = "ARG"
		command = "D+M"
	case "this":
		hackLabel = "THIS"
		command = "D+M"
	case "that":
		hackLabel = "THAT"
		command = "D+M"
	case "temp":
		hackLabel = "R5"
		command = "D+A"
	}
	return hackLabel, command
}

func writePush(segment, indexStr, segLoc, locationCommand string, c *CodeWriter) string {
	contents := ""
	contents += "// push " + segment + " " + indexStr
	switch segment {
	case "constant":
		// get the constant value
		contents += `
    @` + indexStr + `
	D=A`
	case "pointer":
		contents += writePushPointer(indexStr)
	case "static":
		contents += writePushStatic(indexStr, path.Base(c.filepath))
	}
	// For local, argument, this, that, and temp comands
	if segLoc != "" {
		contents += `
	@` + indexStr + `
	D=A
	@` + segLoc + `
	A=` + locationCommand + `
	D=M`
	}
	// push to the stack
	contents += `
	@SP
	A=M
	M=D
	@SP
	M=M+1
`
	return contents
}

func writePop(segment, indexStr, segLoc, locationCommand string, c *CodeWriter) string {
	contents := ""
	contents += "// pop " + segment + " " + indexStr

	// Get segment location and save into R13
	if segLoc != "" {
		contents += saveSegLocInto13(indexStr, segLoc, locationCommand)
	}

	// pop stack into D register
	contents += `
	@SP
	A=M-1
	D=M
	@SP
	M=M-1
`

	if segLoc != "" {
		contents += saveDIntoReg13()
	} else if segment == "pointer" {
		contents += writePopPointer(indexStr)
	} else if segment == "static" {
		contents += writePopStatic(indexStr, c.filepath)
	}
	return contents
}

func getPointerAddr(indexStr string) (contents string) {
	if indexStr == "0" {
		contents += "\n    @R3"
	} else {
		contents += "\n    @R4"
	}
	return contents
}

func writePushPointer(indexStr string) (contents string) {
	contents += getPointerAddr(indexStr)
	contents += `
	D=M`
	return contents
}

func writePopPointer(indexStr string) (contents string) {
	contents += getPointerAddr(indexStr)
	contents += `
	M=D	
`
	return contents
}

func saveSegLocInto13(indexStr, segLoc, locationCommand string) (contents string) {
	contents += `
	@` + indexStr + `
	D=A
	@` + segLoc + `
	D=` + locationCommand + `
	@R13
	M=D`
	return contents
}

func saveDIntoReg13() (contents string) {
	contents += `    @R13
	A=M
	M=D
`
	return contents
}

func writePushStatic(staticVarStr string, filename string) (contents string) {
	filenameClean := filename[0 : len(filename)-len(filepath.Ext(filename))]
	contents += `
	@` + filenameClean + "." + staticVarStr + `
	D=M`

	return contents
}

func writePopStatic(staticVarStr string, filename string) (contents string) {
	filenameClean := filename[0 : len(filename)-len(filepath.Ext(filename))]
	contents += "    @" + filenameClean + "." + staticVarStr + `
	M=D
`
	return contents
}
