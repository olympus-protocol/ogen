package main

import (
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"time"

	"github.com/olympus-protocol/ogen/pkg/burnproof"
)

const MerkleRoot = "f9afaa28423bf0acf296c3ff688a4bbb18e7d0528fd6f2b688028be5614bc386"

var addr string
var proof string
var merkleHash [32]byte

func init() {
	migrationCmd.Flags().StringVar(&addr, "addr", "", "Address used to generate proof to verify against")
	migrationCmd.Flags().StringVar(&proof, "proof", "", "Hex encoded string to verify the coin ownership")

	merkleRootBytes, err := hex.DecodeString(MerkleRoot)
	if err != nil {
		fmt.Println("Unable to decode merkle root bytes")
		os.Exit(0)
	}

	for i, b := range merkleRootBytes[:32/2] {
		merkleHash[i], merkleHash[32-1-i] = merkleRootBytes[32-1-i], b
	}

}

var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Execute a migration proof and returns the serialized proof for broadcasting on ogen",
	Long:  `Execute a migration proof and returns the serialized proof for broadcasting on ogen`,
	Run: func(cmd *cobra.Command, args []string) {
		if addr == "" {
			fmt.Println("Please add an address using the --addr flag to verify the proof against")
			os.Exit(0)
		}

		if proof == "" {
			fmt.Println("Please add your serialized proof data")
			os.Exit(0)
		}

		proofBytes, err := hex.DecodeString(proof)
		if err != nil {
			fmt.Println("Unable to decode proof bytes")
			os.Exit(0)
		}

		t := time.Now()
		if err := burnproof.VerifyBurn(proofBytes, merkleHash, addr); err != nil {
			fmt.Println("Burn verification failed")
			os.Exit(0)
		}
		fmt.Println(time.Since(t))
	},
}

func main() {
	err := migrationCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}
