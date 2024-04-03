package mimir

import "os/exec"

type PathLooker interface {
	LookPath(file string) (string, error)
}

type RealPathLooker struct {
}

func (l *RealPathLooker) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

type FakePathLooker struct{}

func (f *FakePathLooker) LookPath(string) (string, error) {
	return "", nil
}
