package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

const MerkleRoot = "1b58fc1de7722bb17854e3c014a06ce4cf7d34f5298c22a84c55c029dc7c57bf"

func main() {
	out, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	proofBytes, err := base64.StdEncoding.DecodeString(string(out))
	if err != nil {
		panic(err)
	}

	merkleRootHex, err := hex.DecodeString(MerkleRoot)
	if err != nil {
		panic(err)
	}

	merkleRootHash, err := chainhash.NewHash(merkleRootHex)
	if err != nil {
		panic(err)
	}

	t := time.Now()
	if err := burnproof.VerifyBurn(proofBytes, *merkleRootHash, "julian test"); err != nil {
		panic(err)
	}
	fmt.Println(time.Since(t))
}
