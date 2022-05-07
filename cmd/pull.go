/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"

	"github.com/spf13/cobra"

	"ocibundle/internal/action"
)

var pullOpts action.PullOpts

// pull pulls an artifact from OCI registry.
var pullCmd = &cobra.Command{
	Use:           "pull",
	Short:         "pull an artifact from oci registry to a dir with uploaded path",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return action.Pull(context.TODO(), args[0], &pullOpts)
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	flag := pullCmd.Flags()
	flag.StringVarP(&pullOpts.Output, "output", "o", "ocibundle-artifacts", "the output dir path for the artifacts")
	flag.BoolVar(&pullOpts.AllowOverwrite, "overwrite", false, "whether allow overwrite file if it exists")
}
