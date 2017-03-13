package main

import (
	"os"
	"path/filepath"

	"strings"

	"fmt"

	"github.com/jmoravec/VMtranslator/codeWriter"
	"github.com/jmoravec/VMtranslator/parser"
)

func main() {
	file := os.Args[1]

	fileParser := parser.NewParser(file)
	fileParser.OpenFile()
	defer fileParser.Close()

	newFile := strings.TrimSuffix(file, filepath.Ext(file)) + ".asm"

	writer, err := codeWriter.NewCodeWriter(newFile)
	if err != nil {
		fmt.Fprintln(os.Stdout, err.Error())
		panic(err)
	}

	defer writer.Close()
	err = fileParser.Advance()
	for err == nil {
		fileParser.SetCommandType()
		fileParser.SetArg1()
		fileParser.SetArg2()
		switch fileParser.CurrentCommandType {
		case parser.CArithmetic:
			writer.WriteArithmetic(fileParser.Arg1)
		case parser.CPush:
			fallthrough
		case parser.CPop:
			writer.WritePushPop(fileParser.CurrentCommandType, fileParser.Arg1, fileParser.Arg2)
		}
		err = fileParser.Advance()
	}
}
