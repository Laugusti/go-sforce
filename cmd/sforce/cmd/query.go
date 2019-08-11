package cmd

import (
	restapi "github.com/Laugusti/go-sforce/api/rest"
	"github.com/spf13/cobra"
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query [<query>]",
	Args:  cobra.RangeArgs(0, 1),
	Short: "Executes the specified SOQL query",
	Long: `Executes the specified SOQL query.
With no query or when query is -, read standard input`,
	Run: func(cmd *cobra.Command, args []string) {
		// get query from args or stdin
		query := ""
		if len(args) == 1 && args[0] != "-" {
			query = args[0]
		} else {
			query = readAllStdin("Query")
		}

		// create api input
		input := &restapi.QueryInput{
			Query: query,
		}

		// do api request
		out, err := restClient.Query(input)
		exitIfError("Query", err)

		// write result to stdout
		marshalJSONToStdout("Query", out.Result)
	},
}

func init() {
	restCmd.AddCommand(queryCmd)
}
