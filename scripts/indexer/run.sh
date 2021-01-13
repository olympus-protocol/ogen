#!/bin/bash

service postgresql start

./ogen --network testnet --dashboard --rpc_proxy --rpc_proxy_addr 0.0.0.0 & ./ogen indexer testnet --dbconn="postgresql://indexer:indexer@127.0.0.1/indexer"
