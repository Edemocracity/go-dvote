package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"go.vocdoni.io/dvote/client"
	"go.vocdoni.io/dvote/types"
	"nhooyr.io/websocket"
)

var digested bool

var censusCmd = &cobra.Command{
	Use:   "census",
	Short: "census subcommands",
}

var censusAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add [census id] [claim1] ...",
	RunE:  censusAdd,
}

var claimCmd = &cobra.Command{
	Use:   "claim [id] [key]",
	Short: "adds a single claim (hashed pubkey) to a census id",
	RunE:  addClaim,
}

var getRootCmd = &cobra.Command{
	Use:   "root [id]",
	Short: "return the root hash of the census merkle tree",
	RunE:  getRoot,
}

func init() {
	rootCmd.AddCommand(censusCmd)
	censusCmd.AddCommand(censusAddCmd)
	censusCmd.AddCommand(claimCmd)
	censusCmd.AddCommand(getRootCmd)
	censusAddCmd.Flags().BoolVarP(&digested, "digested", "", false, "will digest value in the gateway if false")
}

func censusAdd(cmd *cobra.Command, args []string) error {

	if err := genRandSignKey(); err != nil {
		return err
	}

	if len(args) < 1 {
		return fmt.Errorf("you must provide a census id")
	}

	cl, err := client.New(host)
	if err != nil {
		return err
	}
	defer cl.Conn.Close(websocket.StatusNormalClosure, "")

	var req types.MetaRequest
	req.Method = "addCensus"
	if len(args) > 1 {
		req.PubKeys = args[1:]
	}
	req.CensusID = args[0]
	resp, err := cl.Request(req, signKey)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf(resp.Message)
	}
	fmt.Printf("CensusID: %v\n", resp.CensusID)
	fmt.Printf("URI: %v\n", resp.URI)

	return nil
}

func addClaim(cmd *cobra.Command, args []string) error {

	if err := genRandSignKey(); err != nil {
		return err
	}

	if len(args) < 2 {
		return fmt.Errorf("you must provide a census id and a claim key")
	}

	cl, err := client.New(host)
	if err != nil {
		return err
	}
	defer cl.Conn.Close(websocket.StatusNormalClosure, "")

	var req types.MetaRequest
	req.Method = "addClaim"
	req.CensusID = args[0]
	req.Digested = digested
	req.CensusKey = []byte(args[1])
	resp, err := cl.Request(req, signKey)
	if err != nil {
		return err
	}

	if !resp.Ok {
		return fmt.Errorf(resp.Message)
	}
	fmt.Printf("Root: %v\n", resp.Root)

	return nil
}

func getRoot(cmd *cobra.Command, args []string) error {

	if len(args) < 1 {
		return fmt.Errorf("you must provide a census id")
	}

	cl, err := client.New(host)
	if err != nil {
		return err
	}
	defer cl.Conn.Close(websocket.StatusNormalClosure, "")

	var req types.MetaRequest
	req.Method = "getRoot"
	req.CensusID = args[0]
	resp, err := cl.Request(req, nil)
	if err != nil {
		return err
	}

	if !resp.Ok {
		return fmt.Errorf(resp.Message)
	}
	fmt.Printf("Root: %v", hex.EncodeToString(resp.Root))
	return nil
}
