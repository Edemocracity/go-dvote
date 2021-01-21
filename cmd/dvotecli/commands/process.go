package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"go.vocdoni.io/dvote/client"
	"go.vocdoni.io/dvote/types"
	"go.vocdoni.io/dvote/util"
	"nhooyr.io/websocket"
)

var processCmd = &cobra.Command{
	Use:   "process list|keys|results|finalresults|liveresults",
	Short: "process subcommands",
}

var processListCmd = &cobra.Command{
	Use:   "list [entityId]",
	Short: "list processes of entity",
	RunE:  processList,
}

var processKeysCmd = &cobra.Command{
	Use:   "keys [processId]",
	Short: "list keys of processes",
	RunE:  processKeys,
}

var processResultsCmd = &cobra.Command{
	Use:   "results [processId]",
	Short: "get the results of a process",
	RunE:  getResults,
}

func init() {
	rootCmd.AddCommand(processCmd)
	processCmd.AddCommand(processListCmd)
	processCmd.AddCommand(processKeysCmd)
	processCmd.AddCommand(processResultsCmd)
}

func processList(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("you must provide a entity id")
	}

	cl, err := client.New(host)
	if err != nil {
		return err
	}
	defer cl.Conn.Close(websocket.StatusNormalClosure, "")

	var req types.MetaRequest
	req.Method = "getProcessList"
	req.EntityId, _ = hex.DecodeString(util.TrimHex(args[0]))
	resp, err := cl.Request(req, nil)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf(resp.Message)
	}
	for _, proc := range resp.ProcessList {
		fmt.Printf("%v ", proc)
	}
	return nil
}

func processKeys(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("you must provide a process id")
	}

	cl, err := client.New(host)
	if err != nil {
		return err
	}
	defer cl.Conn.Close(websocket.StatusNormalClosure, "")

	var req types.MetaRequest
	req.Method = "getProcessKeys"
	req.ProcessID, _ = hex.DecodeString(util.TrimHex(args[0]))
	resp, err := cl.Request(req, nil)
	if err != nil {
		return err
	}

	if !resp.Ok {
		return fmt.Errorf(resp.Message)
	}
	if len(resp.EncryptionPublicKeys) == 0 {
		fmt.Print("this is not an encrypted poll")
		return nil
	}

	fmt.Print("Encryption Public Keys: ")
	for _, pubk := range resp.EncryptionPublicKeys {
		fmt.Printf("%v ", pubk.Key)
	}
	fmt.Print("\nCommitment Keys: ")
	for _, cmk := range resp.CommitmentKeys {
		fmt.Printf("%v ", cmk.Key)
	}
	fmt.Print("\nEncryption Private Keys: ")
	for _, pvk := range resp.EncryptionPrivKeys {
		fmt.Printf("%v ", pvk.Key)
	}
	fmt.Print("\nReveal Keys: ")
	for _, rvk := range resp.RevealKeys {
		fmt.Printf("%v ", rvk.Key)
	}
	return nil
}

func getResults(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("you must provide a process id")
	}

	cl, err := client.New(host)
	if err != nil {
		return err
	}
	defer cl.Conn.Close(websocket.StatusNormalClosure, "")

	var req types.MetaRequest
	req.Method = "getResults"
	req.ProcessID, _ = hex.DecodeString(util.TrimHex(args[0]))
	resp, err := cl.Request(req, nil)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf(resp.Message)
	}
	for _, res := range resp.Results {
		fmt.Printf("%v\n", res)
	}
	return nil
}
