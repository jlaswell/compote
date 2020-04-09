/*
Copyright © 2020 John Laswell

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
	"log"
	"os"

	"github.com/jlaswell/compote/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var installCmdShort = "Install packages locked to this project"
var installCmdLong = installCmdShort + `

By default, install requires a composer.lock file to be present
in order to install packages. This prevents the mistake of
updating packages when no lockfile is present.

Examples:
  # Install package locked to this project.
  compote install`

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: installCmdShort,
	Long:  installCmdLong,
	Run:   runInstallCmd,
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().BoolP("no-dev", "", false, "Skip installation of development packages")
	viper.BindPFlag("no-dev", installCmd.Flags().Lookup("no-dev"))
}

func runInstallCmd(cmd *cobra.Command, args []string) {
	filepath := viper.GetString("filepath")

	file, err := pkg.LoadFile(filepath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = pkg.Install(file, viper.GetBool("no-dev"), viper.GetBool("quiet"))
	if err != nil {
		log.Fatal(err)
	}
}
