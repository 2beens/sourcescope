/*
Copyright © 2020 Serj <stubin87@gmail.com>

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

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	BlackPrint  = "\033[1;30m%s\033[0m"
	RedPrint    = "\033[1;31m%s\033[0m"
	GreenPrint  = "\033[1;32m%s\033[0m"
	YellowPrint = "\033[1;33m%s\033[0m"
	WhitePrint  = "\033[1;37m%s\033[0m"
)

var (
	sourceAnalyzer *SourceAnalyzer

	cfgFile          string
	rootDir          string
	importPathPrefix string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "sourcescope",
		Short: "Find the list of packages in need for testing",
		Long:  `Analyze go project for any changes and output the list of packages that require testing, depending on those changes (for GIT repos)`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if helpFlag := rootCmd.Flag("help"); helpFlag != nil && helpFlag.Value.String() == "true" {
		if err := rootCmd.Usage(); err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	if rootDir == "." {
		var err error
		rootDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("cannot get root directory: %s", err.Error())
		}
	}
	if len(rootDir) == 0 {
		log.Fatal("invalid root dir")
	}

	fmt.Print("source path prefix: \t")
	fmt.Printf(RedPrint+"\n", importPathPrefix)
	fmt.Print("root dir: \t\t")
	fmt.Printf(RedPrint+"\n", rootDir)
	fmt.Println()

	sourceAnalyzer = NewSourceAnalyzer(rootDir, importPathPrefix)

	changedPackages, dependentPackages := sourceAnalyzer.GetChangedAndDependentSources()
	dependentPackagesRootFolders := sourceAnalyzer.GetRootFolders(dependentPackages)

	fmt.Printf("\n"+RedPrint+"\n", " >>>>>>>>>>>>>>>>>>>>>>>>>> changed packages:")
	for _, p := range changedPackages {
		fmt.Println("*** " + p)
	}
	fmt.Printf("\n"+RedPrint+"\n", " >>>>>>>>>>>>>>>>>>>>>>>>>> dependent packages:")
	for _, p := range dependentPackages {
		fmt.Println("+++ " + p)
	}
	fmt.Printf("\n"+RedPrint+"\n", " >>>>>>>>>>>>>>>>>>>>>>>>>> dependent packages root folders:")
	for _, f := range dependentPackagesRootFolders {
		fmt.Print("### ")
		fmt.Printf(GreenPrint+"\n", f)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	flags := rootCmd.Flags()
	flags.StringVarP(
		&importPathPrefix,
		"prefix",
		"p",
		"github.com/adjust/backend", // default value
		"import path prefix (e.g. github.com/username/projectname), needed when determining if changed packages are being imported in others",
	)
	flags.StringVarP(
		&rootDir,
		"rootdir",
		"r",
		".", // default value
		"source root directory",
	)
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
