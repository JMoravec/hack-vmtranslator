package codeWriter

import (
	"testing"

	"github.com/jmoravec/VMtranslator/parser"
)

func TestWriteArithmetic(t *testing.T) {
	testCodeWriter, err := NewCodeWriter("test.as")
	if err != nil {
		t.Fatal(err)
	}
	defer testCodeWriter.Close()
	testCodeWriter.WriteArithmetic("eq")
	testCodeWriter.WriteArithmetic("add")
	testCodeWriter.WriteArithmetic("lt")
	testCodeWriter.WriteArithmetic("not")
	testCodeWriter.WriteArithmetic("or")
	testCodeWriter.WritePushPop(parser.CPush, "constant", 99)
}
