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

var diffOpts action.DiffOpts

// diffCmd generates patch pkg from src to dest as the form of OCI-image layout
var diffCmd = &cobra.Command{
	Use:           "diff [target] [source]",
	Short:         "generate patch from source reference to dist reference",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return action.Diff(context.TODO(), args[0], args[1], &diffOpts)
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
	flag := diffCmd.Flags()
	flag.StringVarP(&diffOpts.Output, "output", "o", "ocibundle-patch", "the output dir path for the artifacts")
}
