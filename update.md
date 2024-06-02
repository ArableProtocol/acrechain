
# Update to v1.0.0
#### If you have a node which running on bamboo_9000-1 and installed from https://github.com/ArableProtocol/acrechain/blob/main/install.md
#### You should these steps to update correctly.

### Let's check and run this code
```echo $CHAIN_ID```

### If your result is empty/null, run the code below
```
cat <<EOF >> $HOME/.profile
export CHAIN_ID=bamboo_9000-1
EOF
```
### If your result is bamboo_9000-1, run the code below and update the CHAIN_ID from bamboo_9000-1 to bamboo_9051-1
```
nano $HOME/.profile
```

### Run the code below
```source $HOME/.profile

cd $HOME/acrechain
git checkout v1.0.0
make install
```
```
acred version --long

# name: acre
# server_name: acred
# version: ""
# commit: 01482d6deddda2b0b4a399857857dc2a0dd38555
# build_tags: netgo,ledger
# go: go version go1.18.5 linux/amd64
```

### Copy binary - Setting up config
```
sudo systemctl stop $SERVICE_NAME
sudo cp $HOME/go/bin/$SERVICE_NAME /usr/local/bin/$SERVICE_NAME

$SERVICE_NAME config chain-id $CHAIN_ID
$SERVICE_NAME init $MONIKER_NAME --chain-id $CHAIN_ID -o
```

### Re-create service
```
sudo systemctl disable $SERVICE_NAME
sudo rm /etc/systemd/system/$SERVICE_NAME.service

sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null <<EOF
[Unit]
Description=ACRE Node
After=network-online.target

[Service]
User=$USER
WorkingDirectory=$HOME/.acred
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
wget -O $HOME/.acred/config/genesis.json https://raw.githubusercontent.com/ArableProtocol/acrechain/v1.0.0/networks/bamboo/genesis.json
PEERS="6d41af54405fa98073b178262ec9d083b3f12e67@46.4.81.204:16656,a1900a1eca73c08a2b5718d07d8b649d6e1e0fc9@94.237.27.199:26646,221460c042f6b8314308f4f522b1bdbc15cca1f0@31.7.196.66:26656,cc751411f9e40f8ded410872ea2c73f8261ac0c7@135.181.72.187:11593,ffdd3c4f05b6b080907c40f4f36ac95998dd2c9e@65.109.28.177:49467,79f234d95580c1b49840137131e65a1ce23667cf@65.108.204.119:26616,6f71a4782ad31b24cfcd86f0973708a158ebf05e@65.108.100.214:26656"
sed -i.bak -e "s/^persistent_peers *=.*/persistent_peers = \"$PEERS\"/" $HOME/.acred/config/config.toml
```

### Clear db
```
$SERVICE_NAME tendermint unsafe-reset-all --home $HOME/.acred
```

# Start service
```
sudo systemctl daemon-reload && \
sudo systemctl enable $SERVICE_NAME && \
sudo systemctl restart $SERVICE_NAME && \
sudo journalctl -u $SERVICE_NAME -f -o cat
```
