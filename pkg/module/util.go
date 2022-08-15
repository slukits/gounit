// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package module

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func findModule(dir string) (path, name string, err error) {
	for {
		dd, err := os.ReadDir(dir)
		if err != nil {
			return "", "", err
		}
		for _, d := range dd {
			if d.IsDir() {
				continue
			}
			if d.Name() != "go.mod" {
				continue
			}
			path = dir
			break
		}
		if dir == path {
			break
		}
		if dir == filepath.Dir(dir) {
			break
		}
		dir = filepath.Dir(dir)
	}
	if path == "" {
		return "", "", fmt.Errorf("%w"+"%s", ErrNoModule, dir)
	}
	goMod, err := os.Open(filepath.Join(path, "go.mod"))
	if err != nil {
		return "", "", err
	}
	defer goMod.Close()

	scanner := bufio.NewScanner(goMod)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "module ") {
			continue
		}
		if err := scanner.Err(); err != nil {
			return "", "", err
		}
		name = line[len("module "):]
		break
	}

	return path, name, nil
}
