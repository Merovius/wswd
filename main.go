// Copyright 2023 Axel Wagner
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command wswd assigns working directories to i3 workspaces.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/renameio/v2"
	"go.i3wm.org/i3/v4"
	"golang.org/x/sys/unix"
)

func main() {
	log.SetFlags(log.Lshortfile)
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("usage: %s [-set] <cmd...>", os.Args[0])
	}

	wss, err := i3.GetWorkspaces()
	if err != nil {
		return fmt.Errorf("could not get workspaces: %w", err)
	}
	var cur *i3.Workspace
	for i, ws := range wss {
		if ws.Focused {
			cur = &wss[i]
		}
	}
	if cur == nil {
		return errors.New("no workspace is focused")
	}
	dir, err := cfgDir()
	if err != nil {
		return fmt.Errorf("can not determine config dir: %w", err)
	}
	c := &config{
		dir:        dir,
		cur:        *cur,
		workspaces: wss,
	}
	switch args[0] {
	case "exec":
		args = args[1:]
	case "clean":
		return runClean(ctx, c)
	case "set":
		return runSet(ctx, c)
	case "unset":
		return runUnset(ctx, c)
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
	cmd, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}

	fname := filepath.Join(c.dir, fmt.Sprintf("id-%x", c.cur.ID))
	buf, err := os.ReadFile(fname)
	if err == nil {
		if err := os.Chdir(string(buf)); err != nil {
			return err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	return unix.Exec(cmd, args[1:], os.Environ())
}

type config struct {
	dir        string
	cur        i3.Workspace
	workspaces []i3.Workspace
}

func runSet(ctx context.Context, c *config) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fname := filepath.Join(c.dir, fmt.Sprintf("id-%x", c.cur.ID))
	return renameio.WriteFile(fname, []byte(wd), 0600)
}

func runUnset(ctx context.Context, c *config) error {
	fname := filepath.Join(c.dir, fmt.Sprintf("id-%x", c.cur.ID))
	err := os.Remove(fname)
	if errors.Is(err, fs.ErrNotExist) {
		err = nil
	}
	return err
}

func runClean(ctx context.Context, c *config) error {
	des, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}
	allocated := make(set[string])
	for _, ws := range c.workspaces {
		allocated.Add(fmt.Sprintf("id-%x", ws.ID))
	}
	empty := true
	var errs []error
	for _, de := range des {
		n := de.Name()
		if allocated.Contains(n) {
			empty = false
			continue
		}
		errs = append(errs, os.Remove(filepath.Join(c.dir, n)))
	}
	if empty {
		errs = append(errs, os.Remove(c.dir))
	}
	return errorsJoin(errs...)
}

func cfgDir() (string, error) {
	ucd, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	cfg := filepath.Join(ucd, "wswd")
	if err := os.MkdirAll(cfg, 0700); err != nil {
		return "", err
	}
	return cfg, nil
}

type set[E comparable] map[E]struct{}

func makeSet[E comparable](elements ...E) set[E] {
	s := make(set[E])
	for _, e := range elements {
		s.Add(e)
	}
	return s
}

func (s set[E]) Add(e E) {
	s[e] = struct{}{}
}

func (s set[E]) Contains(e E) bool {
	_, ok := s[e]
	return ok
}

func (s set[E]) Delete(e E) {
	delete(s, e)
}

// TODO: replace with errors.Join once Go 1.20 is released.
func errorsJoin(errs ...error) error {
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	e := &joinError{
		errs: make([]error, 0, n),
	}
	for _, err := range errs {
		if err != nil {
			e.errs = append(e.errs, err)
		}
	}
	return e
}

type joinError struct {
	errs []error
}

func (e *joinError) Error() string {
	var b []byte
	for i, err := range e.errs {
		if i > 0 {
			b = append(b, '\n')
		}
		b = append(b, err.Error()...)
	}
	return string(b)
}

func (e *joinError) Unwrap() []error {
	return e.errs
}
