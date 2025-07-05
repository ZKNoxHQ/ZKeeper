#deploy precomputed tables for falcon field


#deployment script
#!/bin/bash

# Configuration
# Replace with your contract name
#DeployETHFalcon.s.sol, DeployFalcon.s.sol
CONTRACT_NAME="DeployZkK1.s.sol"

# Deploy to networks
echo "Deploying $CONTRACT_NAME with Forge..."
#!/bin/bash

# Configuration
# Replace with your private key   
#PRIVATE_KEY="" 
PUB_KEY="0xB9B0d5C2001AAfb6F3Fb90942F78A5911385cCe8"

#https://explorer.garfield-testnet.zircuit.com/
RPC_ZIRCUIT="https://garfield-testnet.zircuit.com/"

#your APIKEY to verify contract
API_KEY_ETHERSCAN="8S6BUAJMGQ1JU89RHSR4FW772JJW11NTR8"
#selected API KEY
API_KEY=$API_KEY_ETHERSCAN
LINEA_KEY_ETHERSCAN=""
BASE_APIKEY=""
OPTIMISM_APIKEY=""
#selected RPC

RPC=$RPC_ZIRCUIT
# Deploy to networks
echo "RPC used: "$RPC
echo "balance:"

cast balance $PUB_KEY --rpc-url $RPC

#forge script $CONTRACT_NAME --rpc-url $RPC --private-key $PRIVATE_KEY --broadcast --tc Script_Deploy_Falcon --etherscan-api-key $API_KEY --verify --priority-gas-price 1

forge script $CONTRACT_NAME --rpc-url $RPC --ledger --broadcast --tc Script_Deploy_ZkK1 --etherscan-api-key $API_KEY_ETHERSCAN --verify --priority-gas-price 1
