#!/bin/bash

service postgresql start

./ogen --network testnet --dashboard & ./ogen indexer testnet --dbconn="postgresql://indexer:indexer@127.0.0.1/indexer"
