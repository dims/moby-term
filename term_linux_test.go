//+build linux

package term // import "github.com/moby/term"

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gotest.tools/assert"
)

// RequiresRoot skips tests that require root, unless the test.root flag has
// been set
func RequiresRoot(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("skipping test that requires root")
		return
	}
}

func newTtyForTest(t *testing.T) (*os.File, error) {
	RequiresRoot(t)
	file, err := os.OpenFile("/dev/tty", os.O_RDWR, os.ModeDevice)
	if err != nil && strings.Contains(err.Error(), "no such device or address") {
		t.Skip("terminal missing, skipping test")
	}
	return file, err
}

func newTempFile() (*os.File, error) {
	return ioutil.TempFile(os.TempDir(), "temp")
}

func TestGetWinsize(t *testing.T) {
	tty, err := newTtyForTest(t)
	assert.NilError(t, err)
	defer tty.Close()
	winSize, err := GetWinsize(tty.Fd())
	assert.NilError(t, err)
	assert.Assert(t, winSize != nil)

	newSize := Winsize{Width: 200, Height: 200, x: winSize.x, y: winSize.y}
	err = SetWinsize(tty.Fd(), &newSize)
	assert.NilError(t, err)
	winSize, err = GetWinsize(tty.Fd())
	assert.NilError(t, err)
	assert.DeepEqual(t, *winSize, newSize, cmpWinsize)
}

var cmpWinsize = cmp.AllowUnexported(Winsize{})

func TestSetWinsize(t *testing.T) {
	tty, err := newTtyForTest(t)
	assert.NilError(t, err)
	defer tty.Close()
	winSize, err := GetWinsize(tty.Fd())
	assert.NilError(t, err)
	assert.Assert(t, winSize != nil)
	newSize := Winsize{Width: 200, Height: 200, x: winSize.x, y: winSize.y}
	err = SetWinsize(tty.Fd(), &newSize)
	assert.NilError(t, err)
	winSize, err = GetWinsize(tty.Fd())
	assert.NilError(t, err)
	assert.DeepEqual(t, *winSize, newSize, cmpWinsize)
}

func TestGetFdInfo(t *testing.T) {
	tty, err := newTtyForTest(t)
	assert.NilError(t, err)
	defer tty.Close()
	inFd, isTerminal := GetFdInfo(tty)
	assert.Equal(t, inFd, tty.Fd())
	assert.Equal(t, isTerminal, true)
	tmpFile, err := newTempFile()
	assert.NilError(t, err)
	defer tmpFile.Close()
	inFd, isTerminal = GetFdInfo(tmpFile)
	assert.Equal(t, inFd, tmpFile.Fd())
	assert.Equal(t, isTerminal, false)
}

func TestIsTerminal(t *testing.T) {
	tty, err := newTtyForTest(t)
	assert.NilError(t, err)
	defer tty.Close()
	isTerminal := IsTerminal(tty.Fd())
	assert.Equal(t, isTerminal, true)
	tmpFile, err := newTempFile()
	assert.NilError(t, err)
	defer tmpFile.Close()
	isTerminal = IsTerminal(tmpFile.Fd())
	assert.Equal(t, isTerminal, false)
}

func TestSaveState(t *testing.T) {
	tty, err := newTtyForTest(t)
	assert.NilError(t, err)
	defer tty.Close()
	state, err := SaveState(tty.Fd())
	assert.NilError(t, err)
	assert.Assert(t, state != nil)
	tty, err = newTtyForTest(t)
	assert.NilError(t, err)
	defer tty.Close()
	err = RestoreTerminal(tty.Fd(), state)
	assert.NilError(t, err)
}

func TestDisableEcho(t *testing.T) {
	tty, err := newTtyForTest(t)
	assert.NilError(t, err)
	defer tty.Close()
	state, err := SetRawTerminal(tty.Fd())
	defer RestoreTerminal(tty.Fd(), state)
	assert.NilError(t, err)
	assert.Assert(t, state != nil)
	err = DisableEcho(tty.Fd(), state)
	assert.NilError(t, err)
}
