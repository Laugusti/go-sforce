package cmd

import (
	"fmt"

	restapi "github.com/Laugusti/go-sforce/api/rest"
	"github.com/spf13/cobra"
)

// deleteSObjectCmd represents the deleteSObject command
var deleteSObjectCmd = &cobra.Command{
	Use:   "deleteSObject <name> <id>",
	Args:  cobra.ExactArgs(2),
	Short: "Deletes the SObject using the Object Name and Object ID",
	Run: func(cmd *cobra.Command, args []string) {
		// create api input
		input := &restapi.DeleteSObjectInput{
			SObjectName: args[0],
			SObjectID:   args[1],
		}

		// do api request
		_, err := restClient.DeleteSObject(input)
		exitIfError("DeleteSObject", err)

		// print success message
		fmt.Printf("Deleted %s object with Id %q\n", args[0], args[1])
	},
}

func init() {
	restCmd.AddCommand(deleteSObjectCmd)
}
