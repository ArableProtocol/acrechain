# Running a validator

## Hardware Requirements

```
4+ CPU cores
500+GB of SSD disk storage
32+GB of memory (RAM)
100+Mbps network bandwidth
```

Note: As the usage of the blockchain grows, the server requirements would increase.

## Supported OS

We support macOS, Windows and Linux:

```
darwin/arm64
darwin/x86_64
linux/arm64
linux/amd64
windows/x86_64
```

## [Peers](./peers.md)

## [RPC](./rpc.md)


## Install Ubuntu 20.04 on a new server and login as root

## Install ``ufw`` firewall and configure the firewall

```
apt-get update
apt-get install ufw
ufw default deny incoming
ufw default allow outgoing
ufw allow 22
ufw allow 26656
ufw enable
```

## Create a new User

```
# add user
adduser node

# add user to sudoers
usermod -aG sudo node

# login as user
su - node
```

## Install Prerequisites

```
sudo apt update
sudo apt install pkg-config build-essential libssl-dev curl jq git libleveldb-dev -y
sudo apt-get install manpages-dev -y

# install go
curl https://dl.google.com/go/go1.18.5.linux-amd64.tar.gz | sudo tar -C/usr/local -zxvf -

```

```
# Update environment variables to include go
cat <<'EOF' >>$HOME/.profile
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export GO111MODULE=on
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
EOF

```

```
source $HOME/.profile

# check go version
go version

```

## Install ACREChain Node

```
git clone https://github.com/ArableProtocol/acrechain.git
cd acrechain
git checkout v1.1.1
make install
cd
acred version --long
```

You should get the following output:

```
# This section will be updated later once v1.1.1 is released
```

## initialize your Validator node
```
#Choose a name for your validator and use it in place of “<moniker-name>” in the following command:
acred init <moniker-name> --chain-id acre_9052-1

#Example
acred init My_Nodes --chain-id acre_9052-1
```

## Download genesis.json file.
```
wget https://raw.githubusercontent.com/ArableProtocol/acrechain/main/networks/mainnet/acre_9052-1/genesis.json -O $HOME/.acred/config/genesis.json

```

## Update configuration with peers list

```
cd

seeds="ade4d8bc8cbe014af6ebdf3cb7b1e9ad36f412c0@seeds.polkachu.com:12256"

PEERS="ef28f065e24d60df275b06ae9f7fed8ba0823448@46.4.81.204:34656,e29de0ba5c6eb3cc813211887af4e92a71c54204@65.108.1.225:46656,276be584b4a8a3fd9c3ee1e09b7a447a60b201a4@116.203.29.162:26656,e2d029c95a3476a23bad36f98b316b6d04b26001@49.12.33.189:36656,1264ee73a2f40a16c2cbd80c1a824aad7cb082e4@149.102.146.252:26656,dbe9c383a709881f6431242de2d805d6f0f60c9e@65.109.52.156:7656,d01fb8d008cb5f194bc27c054e0246c4357256b3@31.7.196.72:26656,91c0b06f0539348a412e637ebb8208a1acdb71a9@178.162.165.193:21095,bac90a590452337700e0033315e96430d19a3ffa@23.106.238.167:26656"


sed -i.bak -e "s/^persistent_peers *=.*/persistent_peers = \"$PEERS\"/" $HOME/.acred/config/config.toml
```

## Running the validator as a systemd unit
```
cd /etc/systemd/system
sudo nano acred.service
```
Copy the following content into ``acred.service`` and save it.

```
[Unit]
Description=Acred Daemon
#After=network.target
StartLimitInterval=350
StartLimitBurst=10

[Service]
Type=simple
User=node
ExecStart=/home/node/go/bin/acred start
Restart=on-abort
RestartSec=30

[Install]
WantedBy=multi-user.target

[Service]
LimitNOFILE=1048576
```

## Reload the daemon, enable and start the service

```
sudo systemctl daemon-reload
sudo systemctl enable acred

# Start the service
sudo systemctl start acred

# Stop the service
sudo systemctl stop acred

# Restart the service
sudo systemctl restart acred


# For Entire log
journalctl -t acred -o cat

# For Entire log reversed
journalctl -t acred -r -o cat

# Latest and continuous
journalctl -fu acred -o cat
```

## Execute the folloiwng command to get the node id

```
acred tendermint show-node-id
```

## Create a Wallet for your Validator Node

Make sure to copy the 24 words Mnemonics Phrase, save it in a file and store it on a safe location.

```
acred keys add <wallet-name>

#Example
acred keys add my_wallet

```

## Create and Register Your Validator Node
```
acred tx staking create-validator -y \
  --chain-id acre_9052-1 \
  --moniker <moniker-name> \
  --pubkey "$(acred tendermint show-validator)" \
  --amount 5000000000000000000aacre \
  --identity "<Keybase.io ID>" \
  --website "<website-address>" \
  --details "Some description" \
  --from <wallet-name> \
  --commission-rate=0.05 \
  --commission-max-rate=0.20 \
  --commission-max-change-rate=0.01 \
  --min-self-delegation 1
  
#Example

acred tx staking create-validator -y \
  --chain-id acre_9052-1 \
  --moniker "My Node" \
  --pubkey "$(acred tendermint show-validator)" \
  --amount 5000000000000000000aacre \
  --identity "D74433D32938F013" \
  --website "http://www.mywebsite.com" \
  --details "Some description" \
  --from my_wallet \
  --commission-rate=0.05 \
  --commission-max-rate=0.20 \
  --commission-max-change-rate=0.01 \
  --min-self-delegation 1
  
```

## Get Validator Operator Address (Valoper Address)

Make sure to change ``<wallet-name>`` to correct values.

```
acred keys show <wallet-name> --bech val --output json | jq -r .address

#Example
acred keys show my_wallet --bech val --output json | jq -r .address
```

## Delegate ``ACRE`` to Your Node
```
acred tx staking delegate <validator-address> 1000000000000000000aacre --from <wallet-name> --chain-id acre_9052-1 -y

#Example
acred tx staking delegate acrevaloper1y4pfpkwpy6myskp7pne256k6smh2rjtay37kwc 1000000000000000000aacre --from my_wallet --chain-id acre_9052-1 -y

#In the above example we are delegating 1 ACRE to the validator.

```
## Backup Validator node file

Take a backup of the following files after you have created and registered your validator node successfully.

```
/home/node/.acred/config/node_key.json
/home/node/.acred/config/priv_validator_key.json
```

## Withdraw Rewards

Make sure to change ``<validator-operator-address>``, ``<wallet-name>`` to correct values.

```
acred tx distribution withdraw-rewards <validator-address> --from <wallet-name> --chain-id acre_9052-1 -y

#Example
acred tx distribution withdraw-rewards acrevaloper1y4pfpkwpy6myskp7pne256k6smh2rjtay37kwc --from my_wallet --chain-id acre_9052-1 -y
```

## Check Balance of an Address

```
acred query bank balances <wallet-address> --chain-id acre_9052-1

#Example:
acred query bank balances acre1lqx0q6q8qktf0vrgzzzcfjmwkldgav6ztjrexg --chain-id acre_9052-1
```
