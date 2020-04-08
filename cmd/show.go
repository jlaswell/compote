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
package cmd

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jlaswell/compote/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var showCmdShort = "Display information about packages"
var showCmdLong = showCmdShort + `

By default, show will provide basic information specific to the
the currently installed packages. You can toggle more verbose
information using the options available to the show command.

Examples:
  # List basic information for all installed packages.
  compote show

# List information for a specific project locally.
  compote show -f ~/code/jlaswell/my-project`

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: showCmdShort,
	Long:  showCmdLong,
	Run:   runShowCmd,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShowCmd(cmd *cobra.Command, args []string) {
	filepath := viper.GetString("filepath")
	file, err := pkg.LoadFile(filepath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	t := table.NewWriter()
	t.Style().Options = table.OptionsNoBordersAndSeparators
	t.Style().Box.PaddingLeft = ""
	t.Style().Box.PaddingRight = "  "
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"NAME", "VERSION", "DESCRIPTION"})
	for _, p := range file.Dependencies(true) {
		t.AppendRow(table.Row{p.Name, p.Version, p.Description})
	}
	t.Render()
}
