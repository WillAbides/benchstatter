package benchstatter

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/WillAbides/benchstatter/pkg/benchstat"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T, path string) {
	t.Helper()
	ex1 := filepath.Join(path, "ex1.go")
	ex1test := filepath.Join(path, "ex1_test.go")
	err := ioutil.WriteFile(ex1, []byte(ex1Rev1), 0600)
	require.NoError(t, err)
	err = ioutil.WriteFile(ex1test, []byte(ex1Bench), 0600)
	require.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(path, ".gitignore"), []byte("tmp/\n"), 0600)
	require.NoError(t, err)
	mustGit(t, path, "init")
	err = os.MkdirAll(filepath.Join(path, "tmp"), 0700)
	require.NoError(t, err)
	mustGit(t, path, "add", ".")
	mustGit(t, path, "commit", "-m", "initial commit")
	err = ioutil.WriteFile(ex1, []byte(ex1Rev2), 0600)
	require.NoError(t, err)
}

func testInDir(t *testing.T, dir string) {
	t.Helper()
	pwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		t.Helper()
		require.NoError(t, os.Chdir(pwd))
	})
}

func TestStatter_Run(t *testing.T) {
	dir := tmpDir(t)
	setupTestRepo(t, dir)
	testInDir(t, dir)
	statter := Statter{
		BenchCmd:   "go",
		BenchArgs:  "test -bench . -benchmem -count 10 -benchtime 10x .",
		ResultsDir: "./tmp",
		BaseRef:    "HEAD",
		Path:       ".",
		Benchstat:  &benchstat.Benchstat{},
	}
	err := statter.Run()
	require.NoError(t, err)
}

var ex1Rev1 = `
package ex1

import (
	"time"
)

var globalBytes []byte

func doNothing() {
	time.Sleep(10 * time.Millisecond)
	newBytes := []byte("0")
	globalBytes = append(globalBytes, newBytes...)
}
`

var ex1Rev2 = `
package ex1

import (
	"time"
)

var globalBytes []byte

func doNothing() {
	time.Sleep(1 * time.Millisecond)
	newBytes := []byte("1123456789")
	globalBytes = append(globalBytes, newBytes...)
}
`

var ex1Bench = `
package ex1

import (
	"testing"
)

func BenchmarkDoNothing(b *testing.B) {
	globalBytes = []byte{}
	for i := 0; i < b.N; i++ {
		doNothing()
	}
}
`
