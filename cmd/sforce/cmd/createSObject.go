package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	restapi "github.com/Laugusti/go-sforce/api/rest"
	"github.com/spf13/cobra"
)

// createSObjectCmd represents the createSObject command
var createSObjectCmd = &cobra.Command{
	Use:   "createSObject <name> [<file>]",
	Args:  cobra.RangeArgs(1, 2),
	Short: "Creates a new SObject using Object Name and data file",
	Long: `Creates a new SObject using Object Name and data file.
With no file or when file is -, read standard input.`,
	Run: func(cmd *cobra.Command, args []string) {
		// get reader (stdin or file)
		var r io.Reader = os.Stdin
		if len(args) == 2 && args[1] != "-" {
			b, err := ioutil.ReadFile(args[1])
			//if err == os.ErrNotExist{}
			exitIfError("CreateSObject", err)
			r = bytes.NewReader(b)
		}
		// unmarshal to SObject
		var sobj restapi.SObject
		exitIfError("CreateSObject", json.NewDecoder(r).Decode(&sobj))

		// do api request
		input := &restapi.CreateSObjectInput{
			SObjectName: args[0],
			SObject:     sobj,
		}
		out, err := restClient.CreateSObject(input)
		exitIfError("CreateSObject", err)

		// write response to stdout
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "\t")
		exitIfError("CreateSObject", enc.Encode(out.Result))
	},
}

func init() {
	restCmd.AddCommand(createSObjectCmd)
}
