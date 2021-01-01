#!/bin/bash

service postgresql start

./ogen --network testnet --dashboard --rpc_proxy & ./ogen indexer testnet --dbconn="postgresql://indexer:indexer@127.0.0.1/indexer"
