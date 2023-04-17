package interpreter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"io"
	"os"
	"testing"
)

func TestInterpreterSimpleProgram(t *testing.T) {
	i := New(zap.S())
	testProgram := `print "one";
	print true;
	print 2 + 1;`
	expectedOutput := "one\ntrue\n3\n"

	realStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := i.Run(testProgram)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = realStdout

	require.Nil(t, err)
	assert.Equal(t, expectedOutput, string(out), "unexpected output results from program %s", testProgram)
}
