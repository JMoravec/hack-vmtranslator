package parser

import (
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	var tests = []string{
		"bla",
		"/home/joshua",
		"/dev/null",
	}

	for _, testFilename := range tests {
		testParse := NewParser(testFilename)
		if testParse.filepath != testFilename || testParse.currentCommandType != -1 {
			t.Error(
				"For filepath:", testParse.filepath, " and currentCommandType:", testParse.currentCommandType,
				"expected: ", testFilename, " and -1",
			)
		}
	}
}

type openFileHelper struct {
	filepath      string
	expectedError string
}

func TestOpenFile(t *testing.T) {
	newFilename := "testFile.vm"
	var tests = []openFileHelper{
		{"", "Filepath not set"},
		{"/dev/null", "File must have a .vm ending"},
		{"./test.go", "File must have a .vm ending"},
		{"./test.c", "File must have a .vm ending"},
		{"./doesntexist.vm", "no such file or directory"},
		{newFilename, ""},
	}

	// setup test file
	file, err := os.Create(newFilename)
	if err != nil {
		file.Close()
		t.Fatal(newFilename, " was not created")
	}
	file.Close()

	for _, testPair := range tests {
		testParse := &Parser{filepath: testPair.filepath}
		err := testParse.OpenFile()
		if testPair.expectedError != "" && !strings.Contains(err.Error(), testPair.expectedError) {
			t.Error("For ", testPair.filepath,
				"Expected (substr): ", testPair.expectedError,
				"Actual: ", err.Error())
		} else if testPair.expectedError == "" {
			if err != nil {
				t.Error("For", testPair.filepath,
					"Expected:", "No error reading file",
					"Actual:", err.Error())
			}
		}
	}

	// remove generated file
	err = os.Remove(newFilename)
	if err != nil {
		t.Error("Problem removing file", err)
	}
}

type advanceHelper struct {
	inputFilepath string
	endError      string
}

func TestAdvance(t *testing.T) {
	var tests = []advanceHelper{
		{"Badfile.vm", "File has not been opened"},
		{"SimpleAdd.vm", "EOF"},
	}

	for _, test := range tests {
		testParser := NewParser(test.inputFilepath)
		err := testParser.Advance()
		if !strings.Contains(err.Error(), "File has not been opened") {
			t.Error("Didn't get the correct error when attempting to advance without opening the file")
		}
		testParser.OpenFile()
		err = nil

		for err == nil {
			err = testParser.Advance()
		}
		if err.Error() != test.endError {
			t.Error("For: ", test.inputFilepath, "expected:", test.endError, "actual:", err)
		}
	}
}

type isCommentHelper struct {
	input          string
	expectedOutput bool
}

func TestIsCommentOrEmptyLine(t *testing.T) {
	var tests = []isCommentHelper{
		{"              ", true},
		{"//", true},
		{"push 1", false},
		{"     push 1", false},
		{"     //push 1", true},
		{"     /      /", false},
		{"     /   a  /", false},
		{"push 1 // test comment", false},
		{"", true},
	}

	for _, test := range tests {
		actualOutput := isCommentOrEmptyLine(test.input)
		if actualOutput != test.expectedOutput {
			t.Error("For: ", test.input, "Expected: ", test.expectedOutput, "Actual:", actualOutput)
		}
	}
}

type setCommandTypeHelper struct {
	inputLine    string
	expectedType int
}

func TestSetCommandType(t *testing.T) {
	var tests = []setCommandTypeHelper{
		{"push constant 7", CPush},
		{"pop constant 8", CPop},
		{"add", CArithmetic},
		{"sub", CArithmetic},
		{"neg", CArithmetic},
		{"eq", CArithmetic},
		{"gt", CArithmetic},
		{"lt", CArithmetic},
		{"and", CArithmetic},
		{"or", CArithmetic},
		{"not", CArithmetic},
	}

	testParser := &Parser{currentLine: ""}
	err := testParser.SetCommandType()
	if err.Error() != "Current line hasn't been set" {
		t.Error("Blank command did not return the expected error")
	}

	for _, test := range tests {
		testParser.currentCommandType = -1
		testParser.currentLine = test.inputLine

		err := testParser.SetCommandType()
		if err != nil {
			t.Error("For", test.inputLine, "expected no error but got", err)
		}

		if testParser.currentCommandType != test.expectedType {
			t.Error("For", test.inputLine, "Expected:", test.expectedType, "Actual:", testParser.currentCommandType)
		}
	}
}

type setArgHelper struct {
	inputLine   string
	expectedArg string
}

func TestSetArg1(t *testing.T) {
	var tests = []setArgHelper{
		{"push constant 7", "constant"},
		{"    push    constant   7", "constant"},
		{"pop constant 7", "constant"},
		{"  pop local 9", "local"},
		{"add", "add"},
		{"   sub", "sub"},
	}

	testParser := &Parser{currentCommandType: -1}

	for _, test := range tests {
		testParser.currentCommandType = -1
		testParser.currentLine = test.inputLine
		testParser.arg1 = ""
		testParser.splitLine = nil
		err := testParser.SetArg1()
		if err != nil {
			t.Error("Input:", test.inputLine, "Received error:", err)
		} else if testParser.arg1 != test.expectedArg {
			t.Error("Input:", test.inputLine, "Expected:", test.expectedArg, "Actual:", testParser.arg1)
		}
	}
}

type setArg2Helper struct {
	inputLine     string
	expectedValue int
}

func TestSetArg2(t *testing.T) {
	var tests = []setArg2Helper{
		{"push constant 7", 7},
		{"    push    constant   9", 9},
		{"pop constant 70", 70},
		{"  pop local 1", 1},
		{"add", -1},
		{"   sub", -1},
	}

	testParser := &Parser{currentCommandType: -1}
	for _, test := range tests {
		testParser.currentCommandType = -1
		testParser.currentLine = test.inputLine
		testParser.arg1 = ""
		testParser.arg2 = -10
		testParser.splitLine = nil

		err := testParser.SetArg2()
		if test.expectedValue == -1 && err.Error() != "Current command is CArithmetic, no arg2 exists" {
			t.Error("Input:", test.inputLine, "Expected error but did not receive it")
		} else if test.expectedValue != -1 && err != nil {
			t.Error("Input:", test.inputLine, "Received error:", err)
		} else if test.expectedValue != -1 && testParser.arg2 != test.expectedValue {
			t.Error("Input:", test.inputLine, "Expected:", test.expectedValue, "Actual:", testParser.arg1)
		}
	}
}
