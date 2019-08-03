package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	usernameCfgName     = "SFORCE_USERNAME"
	passwdCfgName       = "SFORCE_PASSWORD"
	clientIDCfgName     = "SFORCE_CLIENT_ID"
	clientSecretCfgName = "SFORCE_CLIENT_SECRET"

	loginURLCfgName   = "SFORCE_LOGIN_URL"
	apiVersionCfgName = "SFORCE_API_VERSION"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure the CLI options.",
	Long: `Configure the CLI options. You will be prompted for configuration values
such as your username and password. If your config files does not exists, the sforce CLI
will create it for you (default location ~/.sforce/config.yml). To keep an existing value
, hit enter when prompted for the value. When you are prompted for information, the current 
value will be displayed in [brackets]. Note that the configure command only works with
values from the config file. It does not use any configuration values from environment
variables.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// get current values
		username := credsViper.GetString(usernameCfgName)
		password := credsViper.GetString(passwdCfgName)
		clientID := credsViper.GetString(clientIDCfgName)
		clientSecret := credsViper.GetString(clientSecretCfgName)

		loginURL := configViper.GetString(loginURLCfgName)
		apiVersion := configViper.GetString(apiVersionCfgName)

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
		// get login url
		loginURL, err = getFromUser("Login URL", loginURL, false)
		if err != nil {
			return fmt.Errorf("failed to read login url: %v", err)
		}
		// get api version
		apiVersion, err = getFromUser("API Version", apiVersion, false)
		if err != nil {
			return fmt.Errorf("failed to read login url: %v", err)
		}

		// set credentials
		credsViper.Set(usernameCfgName, username)
		credsViper.Set(passwdCfgName, password)
		credsViper.Set(clientIDCfgName, clientID)
		credsViper.Set(clientSecretCfgName, clientSecret)
		// set config
		configViper.Set(loginURLCfgName, loginURL)
		configViper.Set(apiVersionCfgName, apiVersion)

		// create file if not exist
		if err := createDefaultFileIfNotExits(credsViper, "credentials.yml"); err != nil {
			return err
		}
		if err := createDefaultFileIfNotExits(configViper, "config.yml"); err != nil {
			return err
		}

		// write config to file
		if err := credsViper.WriteConfig(); err != nil {
			return err
		}
		return configViper.WriteConfig()
	},
}

// getFromUser asks user for input and returns the input string.
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

// readInput reads input from stdin.
func readInput() (string, error) {
	s, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

// readSecret reads input from stdin but does not echo.
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
}
