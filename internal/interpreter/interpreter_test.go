package interpreter

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
	assert.Equal(t, expectedOut, output, "tried running program: `%s`", program)
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
	program := `var x = 5;
	var y = 6;
	print y +x;
	print y*y;
	y = x / 3;
	print y;`
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

func TestSimpleScope(t *testing.T) {
	program := `var a = 10; { print a; var a = 20; print a; }`
	expectedOut := "10\n20"
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

func TestClassInstanceField(t *testing.T) {
	program := `class A {}
var a = A();
a.x = 1;
print a.x;
a.x = a.x + 2;
print a.x;`
	expectedOut := "1\n3"
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}

func TestRecursiveFunction(t *testing.T) {
	program := `
fun count(n) {   
	if (n > 1) 
		count(n - 1);
	print n; 
}
count(3);`
	expectedOut := "1\n2\n3"
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
	program := `{
	var x;
	x = 4;
	return x;
	var pointless;
}`
	out, err := testSimpleProgram(program)
	require.NotNil(t, err)
	assert.ErrorContains(t, err, "detected return statement from global scope")
	assert.Empty(t, out)
}

func TestErrorMessageLineNumber(t *testing.T) {
	program := `{
	var x;
	x = 4
	var pointless;
}`
	out, err := testSimpleProgram(program)
	require.NotNil(t, err)
	assert.EqualError(t, err, "failed to parse line, err='Expected ; after Statement. Line 3. Token 'x''")
	assert.Empty(t, out)
}

func TestGlobalVarsInFunc(t *testing.T) {
	program := `
var i = 0;
fun inc() {
  i = i + 1;
}

print i;
inc();
print i;
`
	out, err := testSimpleProgram(program)
	require.Nil(t, err)
	assert.Equal(t, "0\n1", out)
}

func TestFuncPtrReturn(t *testing.T) {
	program := `
fun getNest() {
  fun nested() {
    print "I am nested";
  }
  return nested;
}

var x = getNest();
print x;
x();
`
	out, err := testSimpleProgram(program)
	require.Nil(t, err)
	assert.Equal(t, "<fn nested>\nI am nested", out)
}

func TestBlockScopes(t *testing.T) {
	program := `
{ 
	var x = 4;
	{
		var y = 5;
		print  x + y; // 9
		var x = 6;
		print x; // 6
		x = x + 1;
		print x; // 7
		{
			print x + y; // 12
		}
	}
	print x; // 4
}
`
	out, err := testSimpleProgram(program)
	require.Nil(t, err)
	assert.Equal(t, "9\n6\n7\n12\n4", out)
}

func TestClosureProgram(t *testing.T) {
	program := `
fun makeCounter() {
  var i = 0;
  fun count() {
    i = i + 1;
    print i;
  }

  return count;
}

var counter = makeCounter();
counter(); // "1".
counter(); // "2".
`
	out, err := testSimpleProgram(program)
	require.Nil(t, err)
	assert.Equal(t, "1\n2", out)
}

func TestMultipleSameDeclarationsOutsideOfGlobalScope(t *testing.T) {
	program := `fun bad() {
  var a = "first";
  var a = "second";
}`
	out, err := testSimpleProgram(program)
	require.ErrorContains(t, err, "already exists a variable with name='a' in scope")
	require.Empty(t, out)
}

func TestReturnFromGlobalScope(t *testing.T) {
	program := "return 42;"
	out, err := testSimpleProgram(program)
	require.ErrorContains(t, err, "detected return statement from global scope - not allowed")
	require.Empty(t, out)
}

func TestClassInstanceMethod(t *testing.T) {
	program := `class Bacon {
		eat() {
			print "Crunch crunch crunch!";
		}
	}
	Bacon().eat();`
	expectedOut := "Crunch crunch crunch!"
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}

func TestClassThisUsage(t *testing.T) {
	program := `class Cake {
  taste() {
    var adjective = "delicious";
    print "The " + this.flavor + " cake is " + adjective + "!";
  }
}

var cake = Cake();
cake.flavor = "German chocolate";
cake.taste(); // Prints "The German chocolate cake is delicious!".
`
	out, err := testSimpleProgram(program)
	require.Nil(t, err)
	assert.Equal(t, "The German chocolate cake is delicious!", out)
}

func TestClassInitializer(t *testing.T) {
	program := `class Thing {
  init(){
    this.x = 45;
  }
}
print Thing().x;`
	expectedOut := "45"
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}

func TestClassInitializerWithReturn(t *testing.T) {
	program := `class Thing {
		init() {
			return;
		}
	}
	print Thing();`
	expectedOut := "Thing instance"
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}

func TestClassInheritance(t *testing.T) {
	program := `class Doughnut {
  cook() {
    print "Fry until golden brown.";
  }
}

class BostonCream < Doughnut {}

BostonCream().cook();`
	expectedOut := "Fry until golden brown."
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}

func TestSuperUsage(t *testing.T) {
	program := `class A {
  method() {
    print "A method";
  }
}

class B < A {
  method() {
    print "B method";
  }

  test() {
    super.method();
  }
}

class C < B {}

C().test();`
	expectedOut := "A method"
	testSimpleProgramWorksWithOutput(t, program, expectedOut)
}
