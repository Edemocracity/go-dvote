package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.vocdoni.io/dvote/client"
	"go.vocdoni.io/dvote/types"
	"nhooyr.io/websocket"
)

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "block subcommands",
}

var blockHeightCmd = &cobra.Command{
	Use:   "height",
	Short: "get the current block height",
	RunE:  blockHeight,
}

var blockStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "get the average block time  and timestamp",
	RunE:  blockTime,
}

func init() {
	rootCmd.AddCommand(blockCmd)
	blockCmd.AddCommand(blockHeightCmd)
	blockCmd.AddCommand(blockStatusCmd)
}

func blockHeight(cmd *cobra.Command, args []string) error {
	cl, err := client.New(host)
	if err != nil {
		return err
	}
	defer cl.Conn.Close(websocket.StatusNormalClosure, "")

	block, err := cl.GetCurrentBlock()
	if err != nil {
		return err
	}
	fmt.Printf("Height: %d\n", block)
	return nil
}

func blockTime(cmd *cobra.Command, args []string) error {
	cl, err := client.New(host)
	if err != nil {
		return err
	}
	defer cl.Conn.Close(websocket.StatusNormalClosure, "")

	var req types.MetaRequest
	req.Method = "getBlockStatus"
	resp, err := cl.Request(req, signKey)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf(resp.Message)
	}
	if len(resp.BlockTime) >= 5 {
		fmt.Print("Block Time Average: ")
		fmt.Printf("1m %vms, ", resp.BlockTime[0])
		fmt.Printf("10m %vms, ", resp.BlockTime[1])
		fmt.Printf("1h %vms, ", resp.BlockTime[2])
		fmt.Printf("6h %vms, ", resp.BlockTime[3])
		fmt.Printf("24h %vms\n", resp.BlockTime[4])
	}
	fmt.Printf("Block Timestamp: %v\n", resp.BlockTimestamp)
	return nil
}
