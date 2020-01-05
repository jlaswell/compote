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

To force the installation of packages when missing a lockfile,
the --force option is available.

Examples:
  # Install package locked to this project.
  compote install`

// @todo
//  # Install packages locked to this project, but fallback to the
// # composer.json if composer.lock is missing.
// compote install --force
//
// # Install package from a unique dependency filename.
// compote install -f compote.json --force`

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: installCmdShort,
	Long:  installCmdLong,
	Run:   runInstallCmd,
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().BoolP("no-autoloader", "", false, "Skip autoloading after install")
	// @todo add force when autoloading is more stable
	// installCmd.Flags().BoolP("force", "", false, "Fallback to use composer.json when missing .lock file")
	installCmd.Flags().BoolP("no-dev", "", false, "Skip installation of development packages")
	viper.BindPFlag("no-autoloader", installCmd.Flags().Lookup("no-autoloader"))
	// viper.BindPFlag("force", installCmd.Flags().Lookup("force"))
	viper.BindPFlag("no-dev", installCmd.Flags().Lookup("no-dev"))
}

func runInstallCmd(cmd *cobra.Command, args []string) {
	filepath := viper.GetString("filepath")
	force := viper.GetBool("force")

	file, err := pkg.LoadFile(filepath, force)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = pkg.InstallFile(file, viper.GetBool("no-dev"), viper.GetBool("quiet"))
	if err != nil {
		log.Fatal(err)
	}

	// @todo defer to 'composer dump-auto' until we have implemented autoloading
	// if !viper.GetBool("no-autoloader") {
	// pkg.Autoload()
	// }
}
