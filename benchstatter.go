package benchstatter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Benchstat interface {
	Run(files ...string) error
}

type Statter struct {
	BenchCmd   string
	BenchArgs  string
	ResultsDir string
	BaseRef    string
	Path       string
	Writer     io.Writer
	Benchstat  Benchstat
	Force      bool
}

func (c *Statter) baseOutputFile() (string, error) {
	runner := &gitRunner{
		repoPath: c.Path,
	}
	revision, err := runner.run("rev-parse", c.BaseRef)
	if err != nil {
		return "", err
	}
	revision = bytes.TrimSpace(revision)
	name := fmt.Sprintf("benchstatter-%s.out", string(revision))
	return filepath.Join(c.ResultsDir, name), nil
}

type runBenchmarksResults struct {
	worktreeOutputFile, baseOutputFile string
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

func (c *Statter) runBenchmarks() (*runBenchmarksResults, error) {
	worktreeFilename := filepath.Join(c.ResultsDir, "benchstatter-worktree.out")
	worktreeFile, err := os.Create(worktreeFilename)
	if err != nil {
		return nil, err
	}
	defer worktreeFile.Close()

	cmd := exec.Command(c.BenchCmd, strings.Fields(c.BenchArgs)...)
	fmt.Println(c.BenchArgs)
	cmd.Stdout = worktreeFile
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	baseFilename, err := c.baseOutputFile()
	if err != nil {
		return nil, err
	}

	result := &runBenchmarksResults{
		worktreeOutputFile: worktreeFilename,
		baseOutputFile:     baseFilename,
	}

	if fileExists(baseFilename) && !c.Force {
		return result, nil
	}

	baseFile, err := os.Create(baseFilename)
	if err != nil {
		return nil, err
	}
	defer baseFile.Close()

	baseCmd := exec.Command(c.BenchCmd, strings.Fields(c.BenchArgs)...)
	baseCmd.Stdout = baseFile
	var baseCmdErr error
	runner := &refRunner{
		ref: c.BaseRef,
		gitRunner: gitRunner{
			repoPath:      c.Path,
			gitExecutable: "",
		},
	}
	err = runner.run(func() {
		baseCmdErr = baseCmd.Run()
	})
	if err != nil {
		return nil, err
	}

	if baseCmdErr != nil {
		return nil, err
	}

	return result, nil
}

func (c *Statter) Run() error {
	res, err := c.runBenchmarks()
	if err != nil {
		return err
	}
	return c.Benchstat.Run(res.baseOutputFile, res.worktreeOutputFile)
}
