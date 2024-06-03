# Update the system

```
sudo apt-get update -y && sudo apt upgrade -y
```

# Install git, gcc and make

```
sudo apt-get install make build-essential gcc git jq chrony -y
```

# Install Go

```
wget https://golang.org/dl/go1.18.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.18.5.linux-amd64.tar.gz
```

# Export environment variables

### Please don't forget to define your KEY_NAME and MONIKER_NAME for own at the rows of the end

```
cat <<EOF >> $HOME/.profile
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export GO111MODULE=on
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export CHAIN_ID=bamboo_9051-1
export SERVICE_NAME=acred
export PROJECT_PATH=.acred
export PROJECT_NAME=acred
export TOKEN=aacre
export KEY_NAME=write_your_key_name
export MONIKER_NAME=write_your_moniker_name
EOF
```

```
source $HOME/.profile

go version
# Output should be: go version go1.18.5 linux/amd64
```

# Build

```
git clone https://github.com/ArableProtocol/acrechain && cd acrechain
git checkout v1.0.0
make install

```

```
acred version --long

# name: acre
# server_name: acred
# version: 1.0.0
# commit: 1720aef574485731c27f9186725a26d615be24f4
# build_tags: netgo,ledger
# go: go version go1.18.5 linux/amd64
```

# Copy binary - Setting up config

```
sudo cp $HOME/go/bin/$SERVICE_NAME /usr/local/bin/$SERVICE_NAME


$SERVICE_NAME config chain-id $CHAIN_ID
$SERVICE_NAME config keyring-backend test
$SERVICE_NAME init $MONIKER_NAME --chain-id $CHAIN_ID
```

# Create service

```
sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null <<EOF
[Unit]
Description=$PROJECT_NAME Node
After=network-online.target

[Service]
User=$USER
WorkingDirectory=$HOME/$PROJECT_PATH
ExecStart=$(which $SERVICE_NAME) start
Restart=always
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF
```

# Download genesis

```
wget -O $HOME/$PROJECT_PATH/config/genesis.json https://raw.githubusercontent.com/ArableProtocol/acrechain/v1.0.0/networks/bamboo/genesis.json

PEERS="6d41af54405fa98073b178262ec9d083b3f12e67@46.4.81.204:16656,a1900a1eca73c08a2b5718d07d8b649d6e1e0fc9@94.237.27.199:26646,221460c042f6b8314308f4f522b1bdbc15cca1f0@31.7.196.66:26656,cc751411f9e40f8ded410872ea2c73f8261ac0c7@135.181.72.187:11593,ffdd3c4f05b6b080907c40f4f36ac95998dd2c9e@65.109.28.177:49467,79f234d95580c1b49840137131e65a1ce23667cf@65.108.204.119:26616,6f71a4782ad31b24cfcd86f0973708a158ebf05e@65.108.100.214:26656"
sed -i.bak -e "s/^persistent_peers *=.*/persistent_peers = \"$PEERS\"/" $HOME/$PROJECT_PATH/config/config.toml
```

# Clear db

```
$SERVICE_NAME tendermint unsafe-reset-all --home $HOME/$PROJECT_PATH
```

# Start service

```
sudo systemctl daemon-reload && \
sudo systemctl enable $SERVICE_NAME && \
sudo systemctl restart $SERVICE_NAME && \
sudo journalctl -u $SERVICE_NAME -f -o cat
```

# Create a wallet and a validator after syncing

```
$SERVICE_NAME keys add $KEY_NAME
$SERVICE_NAME tx staking create-validator \
  --amount="<AMOUNT>$TOKEN" \
  --pubkey=$($SERVICE_NAME tendermint show-validator) \
  --moniker=$MONIKER_NAME \
  --chain-id=$CHAIN_ID \
  --commission-rate=0.5 \
  --commission-max-rate=0.1 \
  --commission-max-change-rate=0.1 \
  --min-self-delegation=1 \
  --from=$KEY_NAME
```
