#!/bin/bash

./ogen --dashboard --rpc_proxy & ./ogen indexer testnet --dbconn="postgresql://localhost:5432/ogen-indexer"
