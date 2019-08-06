package cmd

import (
	"fmt"
	"os"
	"strings"

	restapi "github.com/Laugusti/go-sforce/api/rest"
	"github.com/Laugusti/go-sforce/sforce/credentials"
	"github.com/Laugusti/go-sforce/sforce/session"
	"github.com/Laugusti/go-sforce/sforce/sforceerr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var restClient *restapi.Client

// restCmd represents the rest command
var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "The rest command uses the Salesforce REST API",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		missing := []string{}
		// get creds
		username := getConfigString(credsViper, usernameCfgName, &missing)
		password := getConfigString(credsViper, passwordCfgName, &missing)
		clientID := getConfigString(credsViper, clientIDCfgName, &missing)
		clientSecret := getConfigString(credsViper, clientSecretCfgName, &missing)
		// error on missing creds
		if len(missing) > 0 {
			return fmt.Errorf("missing credentials: %s."+
				" You can configure by runnning \"sforce configure\"",
				strings.Join(missing, ", "))
		}

		// get config
		loginURL := getConfigString(configViper, loginURLCfgName, &missing)
		apiVersion := getConfigString(configViper, apiVersionCfgName, &missing)
		// error on missing config
		if len(missing) > 0 {
			return fmt.Errorf("missing configuration: %s."+
				" You can configure by running \"sforce configure\"",
				strings.Join(missing, ", "))
		}

		// create rest client
		restClient = restapi.NewClient(session.Must(session.New(loginURL, apiVersion,
			credentials.New(username, password, clientID, clientSecret))))
		return nil
	},
}

func exitIfError(operation string, err error) {
	if err == nil {
		return
	}
	switch err := err.(type) {
	case *session.LoginError:
		fmt.Fprintf(os.Stderr, "Login failed (%s): %s\n", err.ErrorCode, err.Message)
	case *sforceerr.APIError:
		fmt.Fprintf(os.Stderr, "An error occurred (%s) when calling the %s operation: %s\n",
			err.ErrorCode, operation, err.Message)
	default:
		fmt.Fprint(os.Stderr, err)
	}
	os.Exit(1)
}

func getConfigString(v *viper.Viper, cfgName string, missing *[]string) string {
	s := v.GetString(cfgName)
	if s == "" {
		*missing = append(*missing, cfgName)
	}
	return s
}

func init() {
	rootCmd.AddCommand(restCmd)
}
