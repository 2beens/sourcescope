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
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sourcescope",
	Short: "Find the list of packages in need for testing",
	Long: `Analyze backend source for any changes and output the list of packages that require testing, depending on those changes`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}


	// TODO:
	// 1 - get changed files list
	changedFilesList := getChangedFiles()
	// 2 - remove file-name.go from files list
	changedPackages := getChangedPackages(changedFilesList)

	// 3 - iterate all source files of backend
	// 4 - check import - is current changed package used/imported there ?

	fmt.Println("changed packages:")
	fmt.Println(changedPackages)
	fmt.Println("finished ... ")
}

func getChangedFiles() []string {
	var changedFilesListRaw []byte
	var err error

	cmd := exec.Command( "git", "diff", "--name-only", "master")
	if changedFilesListRaw, err = cmd.Output(); err != nil {
		panic(err.Error())
	}

	changedFiles := string(changedFilesListRaw)
	fmt.Println(changedFiles)

	return strings.Split(changedFiles, "\n")
}

func getChangedPackages(changedFiles []string) []string {
	changedPackages := make(map[string]struct{})
	for _, f := range changedFiles {
		changedFileParts := strings.Split(f, "/")
		var sb strings.Builder
		for _, p := range changedFileParts {
			if strings.HasSuffix(p, ".go") {
				continue
			}
			if sb.Len() > 0 {
				sb.WriteString("/")
			}
			sb.WriteString(p)
		}
		changedPackages[sb.String()] = struct{}{}
	}

	var changedPackagesList []string
	for p := range changedPackages {
		changedPackagesList = append(changedPackagesList, p)
	}

	return changedPackagesList
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sourcescope.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".sourcescope" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".sourcescope")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
