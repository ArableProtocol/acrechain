package main

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
)

func writeCSV(path string, records [][]string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	csvWriter := csv.NewWriter(f)
	err = csvWriter.WriteAll(records)
	if err != nil {
		panic(err)
	}
}

// GenesisState is minimum structure to import airdrop accounts
type GenesisState struct {
	AppState AppState `json:"app_state"`
}

// AppState is app state structure for app state
type AppState struct {
	Staking interface{}
}

type Snapshot struct {
	TotalAtomAmount         sdk.Int `json:"total_atom_amount"`
	TotalOsmosAirdropAmount sdk.Int `json:"total_osmo_amount"`
	NumberAccounts          uint64  `json:"num_accounts"`

	Accounts map[string]SnapshotAccount `json:"accounts"`
}

// SnapshotAccount provide fields of snapshot per account
type SnapshotAccount struct {
	Address        string
	StakedBalance  sdk.Int
	AirdropBalance sdk.Int
}

// setCosmosBech32Prefixes set config for cosmos address system
func setCosmosBech32Prefixes() {
	defaultConfig := sdk.NewConfig()
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(defaultConfig.GetBech32AccountAddrPrefix(), defaultConfig.GetBech32AccountPubPrefix())
	config.SetBech32PrefixForValidator(defaultConfig.GetBech32ValidatorAddrPrefix(), defaultConfig.GetBech32ValidatorPubPrefix())
	config.SetBech32PrefixForConsensusNode(defaultConfig.GetBech32ConsensusAddrPrefix(), defaultConfig.GetBech32ConsensusPubPrefix())
}

// Reference codebase: https://github.com/osmosis-labs/osmosis/blob/4223cbaca92f8f8bc2afacbbf88ad8169116f448/cmd/osmosisd/cmd/airdrop.go
// ExportAirdropSnapshotCmd generates a snapshot.csv from a provided chain genesis export.
func ExportAirdropSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-airdrop-snapshot [input-genesis-file] [output-snapshot-csv]",
		Short: "Export a fairdrop snapshot from a provided cosmos-sdk genesis export",
		Long: `Export a fairdrop snapshot from a provided cosmos-sdk genesis export
Example:
	acred export-airdrop-snapshot ./acrechain_snapshot.json ./snapshot.csv
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			codec := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			genesisFile := args[0]
			snapshotOutput := args[1]

			excludeAddrs := make(map[string]bool)
			excludeAddrs["acre1zasg70674vau3zaxh3ygysf8lgscz50al84jww"] = true
			// 200 AFT tokens
			decimal := sdk.NewDec(1000_000_000).Mul(sdk.NewDec(1000_000_000))
			totalAirdrop := sdk.NewDec(200).Mul(decimal)

			// Read genesis file
			genesisJson, err := os.Open(genesisFile)
			if err != nil {
				return err
			}
			defer genesisJson.Close()

			byteValue, _ := ioutil.ReadAll(genesisJson)

			var genState GenesisState

			// setCosmosBech32Prefixes()
			err = json.Unmarshal(byteValue, &genState)
			if err != nil {
				return err
			}

			stakingBytes, err := json.Marshal(genState.AppState.Staking)
			if err != nil {
				panic(err)
			}
			stakingGen := stakingtypes.GenesisState{}
			codec.MustUnmarshalJSON(stakingBytes, &stakingGen)

			// Produce the map of address to total atom balance, both staked and unstaked
			snapshotAccs := make(map[string]SnapshotAccount)
			// Make a map from validator operator address to the  validator type
			validators := make(map[string]stakingtypes.Validator)
			for _, validator := range stakingGen.Validators {
				validators[validator.OperatorAddress] = validator
			}

			for _, delegation := range stakingGen.Delegations {
				address := delegation.DelegatorAddress
				if excludeAddrs[address] == true {
					continue
				}

				acc, ok := snapshotAccs[address]
				if !ok {
					acc = SnapshotAccount{
						Address:       delegation.DelegatorAddress,
						StakedBalance: sdk.ZeroInt(),
					}
				}

				val := validators[delegation.ValidatorAddress]
				stakedAmount := delegation.Shares.MulInt(val.Tokens).Quo(val.DelegatorShares).RoundInt()
				acc.StakedBalance = acc.StakedBalance.Add(stakedAmount)
				snapshotAccs[address] = acc
			}

			totalStake := sdk.NewInt(0)

			for _, acc := range snapshotAccs {
				totalStake = totalStake.Add(acc.StakedBalance)
			}

			// iterate to find Osmo ownership percentage per account
			for address, acc := range snapshotAccs {
				acc.AirdropBalance = totalAirdrop.Mul(acc.StakedBalance.ToDec()).Quo(totalStake.ToDec()).RoundInt()
				snapshotAccs[address] = acc
			}

			csvRecords := [][]string{{"address", "staked", "airdrop"}}
			for address, acc := range snapshotAccs {
				csvRecords = append(csvRecords, []string{
					address,
					acc.StakedBalance.ToDec().Quo(decimal).String(),
					acc.AirdropBalance.ToDec().Quo(decimal).String(),
				})
			}

			writeCSV(snapshotOutput, csvRecords)

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
