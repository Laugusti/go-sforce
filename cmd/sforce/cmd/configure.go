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
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	usernameCfgName = "SFORCE_USERNAME"
	passwdCfgName   = "SFORCE_PASSWORD"
	cidCfgName      = "SFORCE_CLIENT_ID"
	cSecretCfgName  = "SFORCE_CLIENT_SECRET"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure the CLI options.",
	Long: `Configure the CLI options. You will be prompted for configuration values
such as your username and password. If your config files does not exists, the sforce CLI
will create it for you (default location ~/.sforce.yml). To keep an existing value, hit
enter when prompted for the value. When you are prompted for information, the current 
value will be displayed in [brackets]. Note that the configure command only work with
values from the config file It does not use any configuration values from environment
variables.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// get current values
		username := viper.GetString(usernameCfgName)
		password := viper.GetString(passwdCfgName)
		clientID := viper.GetString(cidCfgName)
		clientSecret := viper.GetString(cSecretCfgName)

		// get username
		username, err := getFromUser("Username", username, false)
		if err != nil {
			return fmt.Errorf("failed to read username: %v", err)
		}
		// get password
		password, err = getFromUser("Password", password, true)
		if err != nil {
			return fmt.Errorf("failed to read password: %v", err)
		}
		// get client id
		clientID, err = getFromUser("Client ID", clientID, false)
		if err != nil {
			return fmt.Errorf("failed to read client id: %v", err)
		}
		// get client secret
		clientSecret, err = getFromUser("Client Secret", clientSecret, true)
		if err != nil {
			return fmt.Errorf("failed to read client secret: %v", err)
		}

		// set configs
		viper.Set(usernameCfgName, username)
		viper.Set(passwdCfgName, password)
		viper.Set(cidCfgName, clientID)
		viper.Set(cSecretCfgName, clientSecret)

		// create config file if not exist
		if viper.ConfigFileUsed() == "" {
			home, err := homedir.Dir()
			if err != nil {
				return fmt.Errorf("could not get home directory: %v", err)
			}
			f, err := os.Create(filepath.Join(home, ".sforce.yml"))
			if err != nil {
				return fmt.Errorf("could not create config file: %v", err)
			}
			if err := f.Close(); err != nil {
				return fmt.Errorf("could not close config file: %v", err)
			}
		}
		// write config to file
		return viper.WriteConfig()
	},
}

func getFromUser(name, current string, secret bool) (string, error) {
	value := current
	// mask if secret
	if secret {
		value = strings.Repeat("*", len(current))
	}
	fmt.Printf("%s [%s]: ", name, value)

	var err error
	if secret {
		value, err = readSecret()
	} else {
		value, err = readInput()
	}
	if err != nil {
		return "", err
	}

	if value == "" {
		return current, nil
	}
	return value, nil
}

func readInput() (string, error) {
	s, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

// readSecret does not echo input
func readSecret() (string, error) {
	b, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func init() {
	rootCmd.AddCommand(configureCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
