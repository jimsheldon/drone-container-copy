// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/types"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/gcrane"
	"github.com/google/go-containerregistry/pkg/name"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	Platform  platformValue
	NoClobber bool   `envconfig:"PLUGIN_NOCLOBBER"`
	Insecure  bool   `envconfig:"PLUGIN_INSECURE"`
	Src       string `envconfig:"PLUGIN_SRC"`
	SrcUser   string `envconfig:"PLUGIN_SRC_USER"`
	SrcPass   string `envconfig:"PLUGIN_SRC_PASS"`
	Dst       string `envconfig:"PLUGIN_DST"`
}

var errConfiguration = errors.New("configuration error")

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {

	err := verifyArgs(&args)
	if err != nil {
		return err
	}

	var opts []crane.Option

	err = craneLogin(&args, opts)
	if err != nil {
		return err
	}

	err = craneCopy(&args)
	if err != nil {
		return err
	}

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

	if args.Dst == "" {
		return fmt.Errorf("no registry provided: %w", errConfiguration)
	}

	return nil
}

func craneLogin(args *Args, craneOpts []crane.Option) error {
	srcRef, err := name.ParseReference(args.Src)
	if err != nil {
		return err
	}
	cf, err := config.Load(os.Getenv("DOCKER_CONFIG"))
	if err != nil {
		return err
	}

	creds := cf.GetCredentialsStore(srcRef.Context().Registry.Name())
	if err := creds.Store(types.AuthConfig{
		ServerAddress: srcRef.Context().Registry.Name(),
		Username:      args.SrcUser,
		Password:      args.SrcPass,
	}); err != nil {
		return err
	}

	if err := cf.Save(); err != nil {
		return err
	}
	fmt.Printf("logged in via %s\n", cf.Filename)

	return nil
}

func craneCopy(args *Args) error {
	jobs := runtime.GOMAXPROCS(0)
	platform := &platformValue{}

	var options []crane.Option
	options = append(options, crane.WithPlatform(platform.platform))
	options = append(options, crane.WithAuthFromKeychain(gcrane.Keychain))

	if args.Insecure {
		options = append(options, crane.Insecure)
	}
	options = append(options, crane.WithJobs(jobs), crane.WithNoClobber(args.NoClobber))

	return crane.Copy(args.Src, args.Dst, options...)
}
