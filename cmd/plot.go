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

var plotOpts action.PlotOpts

// plotCmd plot the tree layer for the artifact
var plotCmd = &cobra.Command{
	Use:           "plot [ref]",
	Short:         "plot the descriptor tree for the artifact",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return action.Plot(context.TODO(), args[0], &plotOpts)
	},
}

func init() {
	rootCmd.AddCommand(plotCmd)
	flag := plotCmd.Flags()
	flag.StringVarP(&plotOpts.Output, "output", "o", "graph.dot", "the filename for the graph")
	flag.BoolVar(&plotOpts.HideLayer, "hide-layer", false, "whether show layer as leaf in plot")
}
