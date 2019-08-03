/*
Copyright © 2019

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
	"path/filepath"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	credsCfgFile string
	credsViper   *viper.Viper
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-sforce-cli",
	Short: "go-sforce-cli is a CLI for Salesforce API",
	Long: `A simple CLI for the Salesforce API written in Go.
		Complete documentation is available at https://github.com/Laugusti/go-sforce-cli`,
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
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&credsCfgFile, "credentials", "", "credentials file (default is $HOME/.sforce/credentials.yml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	credsViper = getViper(credsCfgFile, "credentials")

	// If a config file is found, read it in.
	_ = credsViper.ReadInConfig()
}

func getViper(cfgFile, defaultCfgName string) *viper.Viper {
	v := viper.New()
	if cfgFile != "" {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		cfgHome, err := defaultCfgHome()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in ~/sforce directory with name.
		v.AddConfigPath(cfgHome)
		v.SetConfigName(defaultCfgName)
	}
	v.AutomaticEnv() // read in environment variables that match

	return v
}

// createDefaultFileIfNotExits creates the file if no config file is in use
func createDefaultFileIfNotExits(v *viper.Viper, filename string) error {
	if v.ConfigFileUsed() != "" {
		return nil
	}
	// get config home
	cfgHome, err := defaultCfgHome()
	if err != nil {
		return err
	}
	// create config directory
	if err := os.MkdirAll(cfgHome, os.ModePerm); err != nil {
		return fmt.Errorf("couldn't create config directory: %v", err)
	}
	// create config file
	f, err := os.Create(filepath.Join(cfgHome, filename))
	if err != nil {
		return fmt.Errorf("could not create config file: %v", err)
	}
	// close config file
	if err := f.Close(); err != nil {
		return fmt.Errorf("could not close config file: %v", err)
	}
	return nil
}

func defaultCfgHome() (string, error) {
	// get home directory
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %v", err)
	}
	return filepath.Join(home, ".sforce"), nil
}
