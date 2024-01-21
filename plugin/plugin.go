// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/sirupsen/logrus"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	Clobber  bool   `envconfig:"PLUGIN_CLOBBER"`
	Insecure bool   `envconfig:"PLUGIN_INSECURE"`
	Src      string `envconfig:"PLUGIN_SRC"`
	Dst      string `envconfig:"PLUGIN_DST"`
	User     string `envconfig:"PLUGIN_USER"`
	Pass     string `envconfig:"PLUGIN_PASS"`
}

var errConfiguration = errors.New("configuration error")

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {
	err := verifyArgs(&args)
	if err != nil {
		return err
	}

	err = craneCopy(&args)
	if err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	logrus.Infof("successfully copied %q to %q\n", args.Src, args.Dst)

	return nil
}

// verifyArgs verifies arguments
func verifyArgs(args *Args) error {
	if args.Src == "" {
		return fmt.Errorf("no source provided: %w", errConfiguration)
	}

	if args.Dst == "" {
		return fmt.Errorf("no destination provided: %w", errConfiguration)
	}

	if args.User == "" {
		return fmt.Errorf("no username provided: %w", errConfiguration)
	}

	if args.Pass == "" {
		return fmt.Errorf("no password provided: %w", errConfiguration)
	}

	return nil
}

func craneCopy(args *Args) error {
	jobs := runtime.GOMAXPROCS(0)
	platform := &platformValue{}
	var options []crane.Option
	basicAuth := &authn.Basic{
		Username: args.User,
		Password: args.Pass,
	}

	options = append(options, crane.WithJobs(jobs))
	options = append(options, crane.WithAuth(basicAuth))
	options = append(options, crane.WithPlatform(platform.platform))
	options = append(options, crane.WithNoClobber(!args.Clobber))
	if args.Insecure {
		options = append(options, crane.Insecure)
	}

	return crane.Copy(args.Src, args.Dst, options...)
}
