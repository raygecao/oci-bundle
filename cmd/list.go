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
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

var sortBySemVer bool

// listCmd lists all tags of the repositories in the oci registry
var listCmd = &cobra.Command{
	Use:           "list [repository]",
	Short:         "list all tags of the repository",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tags, err := crane.ListTags(args[0], crane.Insecure)
		if err != nil {
			return err
		}
		// listTags isn't in order which violates oci distribution spec
		if sortBySemVer {
			sort.Slice(tags, func(i, j int) bool {
				semVer1, err1 := semver.NewVersion(tags[i])
				semVer2, err2 := semver.NewVersion(tags[j])
				// sort by semver
				if err1 == nil && err2 == nil {
					return semVer1.GreaterThan(semVer2)
				}
				// sort by string
				if err1 != nil && err2 != nil {
					return tags[i] > tags[j]
				}
				return err1 == nil
			})
		} else {
			sort.Slice(tags, func(i, j int) bool {
				return tags[i] > tags[j]
			})
		}
		table := uitable.New()
		table.AddRow("NO.", "REPOSITORY", "NAME")
		for i, tag := range tags {
			table.AddRow(i, args[0], tag)
		}
		fmt.Println(table)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&sortBySemVer, "sort-by-semver", false, "whether sort by semantic version")
}
