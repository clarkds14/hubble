// Copyright 2017-2020 Authors of Hubble
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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cilium/hubble/cmd/completion"
	"github.com/cilium/hubble/cmd/observe"
	"github.com/cilium/hubble/cmd/peer"
	"github.com/cilium/hubble/cmd/status"
	"github.com/cilium/hubble/cmd/version"
	"github.com/cilium/hubble/pkg"
	"github.com/cilium/hubble/pkg/defaults"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// New create a new root command.
func New() *cobra.Command {
	vp := newViper()
	rootCmd := &cobra.Command{
		Use:           "hubble",
		Short:         "CLI",
		Long:          `Hubble is a utility to observe and inspect recent Cilium routed traffic in a cluster.`,
		SilenceErrors: true, // this is being handled in main, no need to duplicate error messages
		SilenceUsage:  true,
		Version:       pkg.Version,
	}

	cobra.OnInitialize(func() {
		if cfg := vp.GetString("config"); cfg != "" { // enable ability to specify config file via flag
			vp.SetConfigFile(cfg)
		}
		// if a config file is found, read it in.
		if err := vp.ReadInConfig(); err == nil && vp.GetBool("debug") {
			fmt.Fprintln(rootCmd.ErrOrStderr(), "Using config file:", vp.ConfigFileUsed())
		}
	})

	flags := rootCmd.PersistentFlags()
	flags.String("config", "", "Config file (default is $HOME/.hubble.yaml)")
	flags.BoolP("debug", "D", false, "Enable debug messages")
	flags.String("server", defaults.GetDefaultSocketPath(), "Address of a Hubble server")
	flags.Duration("timeout", defaults.DefaultDialTimeout, "Hubble server dialing timeout")
	vp.BindPFlags(flags)

	rootCmd.SetErr(os.Stderr)
	rootCmd.SetVersionTemplate("{{with .Name}}{{printf \"%s \" .}}{{end}}{{printf \"v%s\" .Version}}\n")

	rootCmd.AddCommand(
		completion.New(),
		observe.New(vp),
		peer.New(vp),
		status.New(vp),
		version.New(),
	)
	return rootCmd
}

// Execute creates the root command and executes it.
func Execute() error {
	return New().Execute()
}

// newViper creates a new viper instance configured for Hubble.
func newViper() *viper.Viper {
	vp := viper.New()

	// read config from a file
	vp.SetConfigName("config") // name of config file (without extension)
	vp.SetConfigType("yaml")   // useful if the given config file does not have the extension in the name
	vp.AddConfigPath(".")      // look for a config in the working directory first
	if dir, err := os.UserConfigDir(); err == nil {
		vp.AddConfigPath(filepath.Join(dir, "hubble")) // honor user config dir
	}
	if dir, err := os.UserHomeDir(); err == nil {
		vp.AddConfigPath(filepath.Join(dir, ".hubble")) // fallback to home directory
	}

	// read config from environment variables
	vp.SetEnvPrefix("hubble") // env var must start with HUBBLE_
	vp.AutomaticEnv()         // read in environment variables that match
	return vp
}
