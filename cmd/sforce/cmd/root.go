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

	configCfgFile string
	configViper   *viper.Viper
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-sforce-cli",
	Short: "go-sforce-cli is a CLI for Salesforce API",
	Long: `A simple CLI for the Salesforce API written in Go.
 * Complete documentation is available at https://github.com/Laugusti/go-sforce/tree/master/cmd/sforce`,
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
	rootCmd.PersistentFlags().StringVar(&configCfgFile, "config", "", "config file (default is $HOME/.sforce/config.yml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	credsViper = getViper(credsCfgFile, "credentials")
	configViper = getViper(configCfgFile, "config")

	// If a config file is found, read it in.
	_ = credsViper.ReadInConfig()
	_ = configViper.ReadInConfig()
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

		// Search config in ~/.sforce directory with name.
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

// defaultCfgHome returns the default path for the sforce config files (~/.sforce).
func defaultCfgHome() (string, error) {
	// get home directory
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %v", err)
	}
	return filepath.Join(home, ".sforce"), nil
}
