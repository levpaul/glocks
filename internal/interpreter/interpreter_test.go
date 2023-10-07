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

// TODO: allow this function to accept multi-line programs
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

func TestBasicVarsAndArithmetic(t *testing.T) {
	program := `var x = 5; var y = 6; print y +x; print y*y; y = x / 3; print y;`
	expectedOut := "11\n36\n1.6666666666666667"
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}

func TestMissingSemiColon(t *testing.T) {
	_, err := testSimpleProgram(`print "hello world"`)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Expected ; after Statement"))
}

func TestAssigningToUninitializedVarError(t *testing.T) {
	_, err := testSimpleProgram(`x = 5;`)
	require.Error(t, err)
	assert.Equal(t, "failed to evaluate expression: 'attempted to set variable 'x' but does not exist'", err.Error())
}

func TestSimpleWhileLoop(t *testing.T) {
	program := `var x = 0; while ( x < 5 ) { x = x + 1; print x; }`
	expectedOut := "1\n2\n3\n4\n5"
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}

func TestSimpleForLoop(t *testing.T) {
	program := `for (var i = 1; i <= 5; i = i + 1 ){ print i; }`
	expectedOut := "1\n2\n3\n4\n5"
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}
func TestFibonacciFor(t *testing.T) {
	program := `var a = 0; var temp; for (var b = 1; a < 10000; b = temp + b) { print a; temp = a; a = b; }`
	expectedOut := "0\n1\n1\n2\n3\n5\n8\n13\n21\n34\n55\n89\n144\n233\n377\n610\n987\n1597\n2584\n4181\n6765"
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}

func TestRecursiveFunction(t *testing.T) {
	program := `fun countSkip(n) {   if (n > 1) countSkip(n - 2);   print n; }
countSkip(10);`
	expectedOut := "0\n2\n4\n6\n8\n10"
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}

func TestReturnStmt(t *testing.T) {
	program := `fun addTwo(n) { return n + 2; }
print addTwo(1);`
	expectedOut := "3"
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}

func TestUnexpectedLoneReturnStmt(t *testing.T) {
	program := `return 4;`
	out, err := testSimpleProgram(program)
	require.NotNil(t, err)
	assert.Errorf(t, err, "Unexpected 'return' expression found. Expected to be within a function")
	assert.Empty(t, out)
}

func TestUnexpectedReturnStmtInBlock(t *testing.T) {
	program := `{var x; x = 4; return x; var pointless;}`
	out, err := testSimpleProgram(program)
	require.NotNil(t, err)
	assert.Errorf(t, err, "Unexpected 'return' expression found. Expected to be within a function")
	assert.Empty(t, out)
}
