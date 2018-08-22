#!/bin/sh
set -e

mkdir -p /ethdata/keystore
mkdir -p /ethdata/geth
mkdir -p /ethdata/logs

echo '{"address":"7081536b915595489c192d95829332a7c3a1fd64","crypto":{"cipher":"aes-128-ctr","ciphertext":"4ce4ffa095a510f3655a2809bff8647a0197b505e56165a1542ab89279689694","cipherparams":{"iv":"e2fd841bf187b68f10ce5a1b2943c2c9"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"619f9407eab2e5d8d50edfece719bbb9b5cea35d8b6aade3f08f52cc02ff8001"},"mac":"459bcb3f49839d4eb12180380633d53c172beffe9cd55ecccc22f8063b215fa4"},"id":"6abc8ea4-457f-4cd1-8156-f39a4152ae08","version":3}' > /ethdata/keystore/key1
echo 'pas8w0rd' > /password.txt

geth --datadir /ethdata init /genesis.json
geth --metrics --datadir /ethdata --port 30301 --rpcport 8101 --syncmode full --networkid 19 --rpc --rpcaddr 0.0.0.0 --rpcapi debug,db,personal,eth,network,web3,net,miner --unlock 7081536b915595489c192d95829332a7c3a1fd64 --password /password.txt --ipcpath run/geth.ipc --mine
