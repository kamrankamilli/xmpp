// Copyright 2022 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

// Package python facilitates integration testing against Python scripts.
package python // import "kamrankamilli/xmpp/internal/integration/python"

import (
	"context"
	"io"
	"testing"
	"text/template"

	"kamrankamilli/xmpp/internal/attr"
	"kamrankamilli/xmpp/internal/integration"
)

const cmdName = "python"

func getConfig(cmd *integration.Cmd) Config {
	if cmd.Config == nil {
		cmd.Config = Config{}
	}
	return cmd.Config.(Config)
}

func defaultConfig(cmd *integration.Cmd) error {
	tmpl, err := template.New("python").Parse(baseTest)
	if err != nil {
		return err
	}

	err = integration.TempFile(baseFileName, func(cmd *integration.Cmd, w io.Writer) error {
		cfg := getConfig(cmd)
		return tmpl.Execute(w, cfg)
	})(cmd)
	if err != nil {
		return err
	}

	cfg := getConfig(cmd)
	return integration.Args(append([]string{"-m", baseModule}, cfg.Args...)...)(cmd)
}

// Import causes the given script to be written out to the working directory and
// the class name in that script to be imported by the main test runner and
// executed.
func Import(class, script string) integration.Option {
	return func(cmd *integration.Cmd) error {
		fName := "tmp_" + attr.RandomID()
		cfg := getConfig(cmd)
		cfg.Imports = append(cfg.Imports, []string{fName, class})
		cmd.Config = cfg
		return integration.TempFile(fName+".py", func(cmd *integration.Cmd, w io.Writer) error {
			_, err := io.WriteString(w, script)
			return err
		})(cmd)
	}
}

// Args sets additional command line args to be passed to the script (ie. after
// the script name).
// If you want to pass arguments to the python process (before the script name)
// use integration.Args.
func Args(f ...string) integration.Option {
	return func(cmd *integration.Cmd) error {
		cfg := getConfig(cmd)
		cfg.Args = append(cfg.Args, f...)
		cmd.Config = cfg
		return nil
	}
}

// Test starts a Python script and returns a function that runs subtests using
// t.Run.
// Multiple calls to the returned function will result in uniquely named
// subtests.
// When all subtests have completed, the daemon is stopped.
func Test(ctx context.Context, t *testing.T, opts ...integration.Option) integration.SubtestRunner {
	opts = append(opts, defaultConfig)
	return integration.Test(ctx, cmdName, t, opts...)
}
