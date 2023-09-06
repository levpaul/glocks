package interpreter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"io"
	"os"
	"strings"
	"testing"
)

func TestInterpreterSimpleProgram(t *testing.T) {
	program := `print "one";
	print true;
	print 2 + 1;`
	expectedOutput := "one\ntrue\n3"

	testSimpleProgramWorksWithOutput(t, program, expectedOutput)
}

func testSimpleProgram(program string) (string, error) {
	i := New(zap.S())

	realStdout := os.Stdout
	r, w, _ := os.Pipe()
	defer func() {
		os.Stdout = realStdout
		_ = w.Close()
	}()
	os.Stdout = w

	err := i.Run(program)
	_ = w.Close()
	out, _ := io.ReadAll(r)

	return strings.Trim(string(out), "\n"), err
}

func testSimpleProgramWorksWithOutput(t *testing.T, program, expectedOut string) {
	output, err := testSimpleProgram(program)
	require.Nil(t, err, "expected no errors when running program: `%s`")
	assert.Equal(t, output, expectedOut, "tried running program: `%s`", program)
}

func TestAndFunctionality(t *testing.T) {
	program := `var x = true; var y = false; print(x and y);`
	expectedOut := "false"
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}

func TestOrFunctionality(t *testing.T) {
	cases := map[string]string{
		`print true or false;`:                        "true",
		`var x = true; var y = false; print(x or y);`: "true",
		`print 0 or false;`:                           "0",
		`print 1 or false;`:                           "1",
		`print -1 or false;`:                          "-1",
		`print "hi" or false and false;`:              "hi",
		`print ("hi" or false) and false;`:            "false",
	}
	for program, expectedOutput := range cases {
		testSimpleProgramWorksWithOutput(t, program, expectedOutput)
	}
}
