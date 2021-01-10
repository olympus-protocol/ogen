package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/olympus-protocol/ogen/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var sender = "tlpub1a5v7eu8x3ecmj05pf08edzt9k3x89eu4ugacsr"

var receiver = "tlpub1nfq9r4kuzxwc207f4peyx7kz7860vke30wwr8x"

var password = "123"

func main() {
	rpcClient := client("localhost:24127")

	info, err := rpcClient.wallet.OpenWallet(context.Background(), &proto.WalletReference{Name: "sender", Password: password})
	if err != nil {
		panic(err)
	}

	if !info.Success {
		panic(info.Success)
	}

	for {
		hash, err := rpcClient.wallet.SendTransaction(context.Background(), &proto.SendTransactionInfo{
			Account: receiver,
			Amount:  "0.01",
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(hash)
	}

}

type Client struct {
	wallet proto.WalletClient
}

func client(addr string) *Client {
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		panic("unable to connect to rpc server")
	}

	client := &Client{
		wallet: proto.NewWalletClient(conn),
	}

	return client
}
