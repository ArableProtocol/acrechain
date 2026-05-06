package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ArableProtocol/acrechain/cmd/config"
	appparams "github.com/ArableProtocol/acrechain/cmd/config"
	minttypes "github.com/ArableProtocol/acrechain/x/mint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	"github.com/spf13/cobra"
	tmtypes "github.com/tendermint/tendermint/types"
)

// PrepareGenesisCmd returns generate-genesis cobra Command.
func GenerateGenesisCmd(defaultNodeHome string, mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-genesis [chain_id]",
		Short: "Generate a genesis file with initial setup",
		Long: `Generate a genesis file with initial setup.
Example:
	acred generate-genesis acre_9052-1
	- Check input genesis:
		file is at ~/.acred/config/genesis.json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.Codec
			cdc := depCdc
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			// read genesis file
			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// get genesis params
			chainID := args[0]

			// run Prepare Genesis
			appState, genDoc, err = PrepareGenesis(clientCtx, appState, genDoc, chainID)
			if err != nil {
				return err
			}

			// validate genesis state
			if err = mbm.ValidateGenesis(cdc, clientCtx.TxConfig, appState); err != nil {
				return fmt.Errorf("error validating genesis file: %s", err.Error())
			}

			// save genesis
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			err = genutil.ExportGenesisFile(genDoc, genFile)
			return err
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func PrepareGenesis(clientCtx client.Context, appState map[string]json.RawMessage, genDoc *tmtypes.GenesisDoc, chainID string) (map[string]json.RawMessage, *tmtypes.GenesisDoc, error) {
	depCdc := clientCtx.Codec
	cdc := depCdc

	// chain params genesis
	genDoc.ChainID = chainID
	genDoc.GenesisTime = time.Unix(1669042800, 0) // Monday, November 21, 2022 3:00:00 PM GMT+0000
	genDoc.ConsensusParams = tmtypes.DefaultConsensusParams()
	genDoc.ConsensusParams.Block.MaxBytes = 21 * 1024 * 1024
	genDoc.ConsensusParams.Block.MaxGas = 300_000_000

	// mint module genesis
	mintGenState := minttypes.DefaultGenesisState()
	mintGenState.Params = minttypes.DefaultParams()
	mintGenState.Params.MintDenom = appparams.BaseDenom
	mintGenState.Params.MintingRewardsDistributionStartTime = 1671033600 // Wed Dec 14 2022 16:00:00 GMT+0000
	mintGenState.Params.NextRewardsReductionTime = 1676390400            // Tue Feb 14 2023 16:00:00 GMT+0000

	mintGenStateBz, err := cdc.MarshalJSON(mintGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal mint genesis state: %w", err)
	}
	appState[minttypes.ModuleName] = mintGenStateBz

	// bank module genesis
	bankGenState := banktypes.DefaultGenesisState()
	bankGenState.Params = banktypes.DefaultParams()

	bankGenState.Supply = sdk.NewCoins(sdk.NewCoin(appparams.BaseDenom, sdk.NewInt(350_000_000).Mul(config.DecimalReduction))) // 350M ACRE

	genAccounts := []authtypes.GenesisAccount{}

	addrStrategicReserve, err := sdk.AccAddressFromBech32("acre1zasg70674vau3zaxh3ygysf8lgscz50al84jww") // 0x17608f3F5eAB3BC88Ba6Bc48824127fa218151fD
	if err != nil {
		return nil, nil, err
	}
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(addrStrategicReserve, nil, 0, 0))

	// send tokens to genesis validators
	genesisValidators := []string{
		"acre12g66na08jrr8957uv7ftlmt0xw56j8qf0vgknz", // anonstake
		"acre1chtemgelehwh5k7nqv6wpgmmn46z33gttsw4mu", // vanguard
		"acre1kqlnqxf8clakdm8dqlxyjc7qcvvd88m080ftau", // azstake
		"acre1muhcvklnqrjv6eguw8chshm23pwp2zp0gyew92", // b-harvest
		"acre1j9r5gufue4lw3fmws74lmzqex389s6azu86xjx", // bitcat
		"acre18p32srqkjjqax6f99ql9ucwczc2aemmer33yde", // BitNordic
		"acre1dgeda9nt3fqsfwgmf7hm5xpffsefh4x7pptarn", // blockscape
		"acre1nq9utvcs99tqj680670vf60vmpypl0njuvstea", // blockswell
		"acre1f9nrky5rvs96rmsk7te4jesv92qsm25d9h0p0v", // capital magnate
		"acre1asr74pcdvpwkvtsla6tqqmhuyhewtnl5hze86q", // chainofsecrets
		"acre1csn8s4g52upaadcxum4sqnxprtc683ukr5uzk8", // Chrysoprase
		"acre1qknphqq8vkekyw447kdcjjj58qxk36uxtaz9yw", // coinstamp
		"acre1cy2pzmnfwg6njx502uttjrkmuf4hjtjey9suhk", // D-stake
		"acre1mxhx8ar3f2thtntdsrn4sgv68duxvcxq28gw43", // Enigma
		"acre1ddceyf8tcd7jxwrzaryuckwuzjhge93spcglca", // ericet
		"acre1rvrecs7f3pdlplq75nhqtmegruthhsn8lcrndl", // everstake
		"acre1e76jsstf39zanx846vptzwdegcced09as0qjqp", // french chocolatine
		"acre17j2z96kwktql80ql9qa3ljg9zj0jvue5ljzzxt", // GalaxyStaking
		"acre1ws43egqv720nglj53ks2qqv3z6lzrz3a4g9z2n", // AutoStake
		"acre1ygll7dmlsufgs6yarrnm8ljsed4tpafeg2cxu8", // dankuzone
		"acre1tf33ygp6v4pp7ak3jqzx2kg559xk9ln27h0x7t", // frens
		"acre1wdn6p3ngzmqjmk2mzvmagfu7lhmetamcevhkjm", // mandragora
		"acre1qt6yf3fvz250uddj04rxglfd40v8nruq0rh6py", // NodeStake
		"acre12mq0cjukdx3v2texmg07pr9c4v0ca67cqgyfkf", // PFC
		"acre157dds6jpnvzfnnez599d0ukjvgu2r944cnm7hg", // Zenscape
		"acre15avy39shq7cllgvrvuyfm2c4w4kgkgu86dyqhv", // HashQuark
		"acre1fgs9uuxjwrclak88e633nw75j6ckxmdx09v0pw", // highstakes
		"acre1veyw56cksj64737w5exp4x366hqqsarlrvg57m", // HodlGlobal
		"acre1ax3m50pulmjude2nw2emrt79jxjkpwrldwvnej", // Illuminati
		"acre1ke8tvwj0vdk72m80y5rzu97t2fy9wu3npaelug", // InBWet
		"acre14ye7l35ary0musqs4dn6aj79u3lg5gpfxthwaj", // itgold
		"acre16ta5jvmvj90l3zgry98p3gafqnn8e3aq00987r", // jet-node
		"acre1llprqt8prvf364tgggx8as6yxcw46g4mjzsxrm", // Kalia Network
		"acre1zgpquw7zaphz8n4qn6lgstfaac9zxwk8jl8eey", // lux8
		"acre1p4rah2cttcuqffky7gk4a4auv8hhtmmqvv25ha", // Machfund
		"acre106ukr5w6a95kmtgh4xc9a3n5xcw2rpzqwlszv8", // Marionode
		"acre1wevs2p8t2khs3zp9qs3s7cx4qxu2mz6zfjej6q", // Masternode24
		"acre1rals2gachf9555wj5puhm7l4wnkevwjefwwgt5", // Matrixed Link
		"acre1c3qf72rlsxn6ttf9v56a2x8tzwu6dhqkm3yj7u", // MatrixStake
		"acre1dnwj3dgdfaazgrjtjjffq9cvjft3e9lnvmr700", // MindHeartSoul
		"acre1guxlxgr5flga44vvg7skqp0r5y67a5xnree00c", // namdokmai
		"acre1gr0kmjvkgsf8ph5x9y5f0q6e3f9atzcuk5rqss", // noderunners
		"acre1p6rtyjs3dr9eaknsqavskyz7dzjnesjxwmn5qc", // orbital apes
		"acre1aev5mdduh578z5z894kk2cauxqntjfj6hx4s84", // P-OPS
		"acre19jkd68j79mulnx6pgyqgjwatrvhgdl0clxg59n", // QwertySoftware
		"acre1yppufjenpmk2xueazr9zysgh52f7h3rsv206vn", // Ramuchi
		"acre1lj9tn3huf8zm3ncegmefv0rzwmhj3gz7tmdkn2", // silent
		"acre1fac8t87smt4e2j355glhr3qungqk578mkj4agp", // skynet
		"acre1kskw99wharjx3gdgz8afw30rs7h4fgm7ath2re", // stakeordie
		"acre17gc07hawajnfg7e4539pmps0zfkwrdf4tg8t9m", // Stake-Take
		"acre1hwy0kh4g4at7degm9yfrusukv97ua8lhqv6jga", // Stake.Works
		"acre146q22dq5c39nwfrn5dgd8uxjjrmptpvc5agkam", // Staketab
		"acre1y4pfpkwpy6myskp7pne256k6smh2rjtaaf0xv9", // Synergy Nodes
		"acre1tc94c0ljexukfjrf35ukdjr0puutvxmy9v6yjz", // TheNOP
		"acre1x0m2j2xxgmsj2dmyajcx0jmyjzkd0t9n2xr8ca", // ValidatorRun
		"acre16v8tgcl2a72zpg74us65wv7626mcnu22xjslmq", // VaultStaking
		"acre1rl5h39q804mxypn5gyy0a2uzyrqeydg2eqhteh", // Web3ident
		"acre12242xw27r9v2dplr9hhy3ftpgn55tehg73ajk8", // Web34ever
		"acre1xgymy6z8futd73g9y9gdgndqj3d4zspuc43e8a", // Wetez
		"acre1hvn8advpjvhkdldjqfayqz54p6sgm5djf380xy", // windpowerstake
		"acre1sn80um2chzav6mytlfjthlvaf97ta77hfec6ng", // Yurbason
	}

	totalValidatorInitialCoins := sdk.NewCoins()
	validatorInitialCoins := sdk.NewCoins(sdk.NewCoin(appparams.BaseDenom, sdk.NewInt(120).Mul(config.DecimalReduction))) // 120 ACRE
	for _, address := range genesisValidators {
		bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{
			Address: address,
			Coins:   validatorInitialCoins,
		})
		addr, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			return nil, nil, err
		}
		totalValidatorInitialCoins = totalValidatorInitialCoins.Add(validatorInitialCoins...)
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(addr, nil, 0, 0))
	}

	// strategic reserve = supply - validator initial coins
	bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{
		Address: addrStrategicReserve.String(),
		Coins:   bankGenState.Supply.Sub(totalValidatorInitialCoins),
	})

	bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}
	appState[banktypes.ModuleName] = bankGenStateBz

	// account module genesis
	authGenState := authtypes.GetGenesisStateFromAppState(depCdc, appState)
	authGenState.Params = authtypes.DefaultParams()

	accounts, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		panic(err)
	}

	authGenState.Accounts = append(authGenState.Accounts, accounts...)
	authGenStateBz, err := cdc.MarshalJSON(&authGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal staking genesis state: %w", err)
	}
	appState[authtypes.ModuleName] = authGenStateBz

	// staking module genesis
	stakingGenState := stakingtypes.GetGenesisStateFromAppState(depCdc, appState)
	stakingGenState.Params = stakingtypes.DefaultParams()
	stakingGenState.Params.BondDenom = appparams.BaseDenom
	stakingGenStateBz, err := cdc.MarshalJSON(stakingGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal staking genesis state: %w", err)
	}
	appState[stakingtypes.ModuleName] = stakingGenStateBz

	// distribution module genesis
	distributionGenState := distributiontypes.DefaultGenesisState()
	distributionGenState.Params = distributiontypes.DefaultParams()
	distributionGenState.Params.BaseProposerReward = sdk.ZeroDec()
	distributionGenState.Params.BonusProposerReward = sdk.ZeroDec()
	distributionGenState.Params.CommunityTax = sdk.ZeroDec()
	distributionGenState.FeePool.CommunityPool = sdk.NewDecCoinsFromCoins(sdk.NewCoins()...)
	distributionGenStateBz, err := cdc.MarshalJSON(distributionGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal distribution genesis state: %w", err)
	}
	appState[distributiontypes.ModuleName] = distributionGenStateBz

	// gov module genesis
	govGenState := govtypes.DefaultGenesisState()
	defaultGovParams := govtypes.DefaultParams()
	govGenState.DepositParams = defaultGovParams.DepositParams
	govGenState.DepositParams.MinDeposit = sdk.Coins{sdk.NewCoin(appparams.BaseDenom, sdk.NewInt(500).Mul(config.DecimalReduction))} // 500 ACRE
	govGenState.TallyParams = defaultGovParams.TallyParams
	govGenState.VotingParams = defaultGovParams.VotingParams
	govGenState.VotingParams.VotingPeriod = time.Hour * 24 * 2 // 2 days
	govGenStateBz, err := cdc.MarshalJSON(govGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal gov genesis state: %w", err)
	}
	appState[govtypes.ModuleName] = govGenStateBz

	// slashing module genesis
	slashingGenState := slashingtypes.DefaultGenesisState()
	slashingGenState.Params = slashingtypes.DefaultParams()
	slashingGenState.Params.SignedBlocksWindow = 10000
	slashingGenStateBz, err := cdc.MarshalJSON(slashingGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal slashing genesis state: %w", err)
	}
	appState[slashingtypes.ModuleName] = slashingGenStateBz

	// crisis module genesis
	crisisGenState := crisistypes.DefaultGenesisState()
	crisisGenState.ConstantFee = sdk.NewCoin(appparams.BaseDenom, sdk.NewInt(1).Mul(config.DecimalReduction)) // 1 ACRE
	crisisGenStateBz, err := cdc.MarshalJSON(crisisGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal crisis genesis state: %w", err)
	}
	appState[crisistypes.ModuleName] = crisisGenStateBz

	// ica module genesis
	icaGenState := icatypes.DefaultGenesis()
	icaGenState.HostGenesisState.Params.AllowMessages = []string{
		"/cosmos.bank.v1beta1.MsgSend",
		"/cosmos.bank.v1beta1.MsgMultiSend",
		"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress",
		"/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission",
		"/cosmos.distribution.v1beta1.MsgFundCommunityPool",
		"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
		"/cosmos.feegrant.v1beta1.MsgGrantAllowance",
		"/cosmos.feegrant.v1beta1.MsgRevokeAllowance",
		"/cosmos.gov.v1beta1.MsgVoteWeighted",
		"/cosmos.gov.v1beta1.MsgSubmitProposal",
		"/cosmos.gov.v1beta1.MsgDeposit",
		"/cosmos.gov.v1beta1.MsgVote",
		"/cosmos.staking.v1beta1.MsgEditValidator",
		"/cosmos.staking.v1beta1.MsgDelegate",
		"/cosmos.staking.v1beta1.MsgUndelegate",
		"/cosmos.staking.v1beta1.MsgBeginRedelegate",
		"/cosmos.staking.v1beta1.MsgCreateValidator",
		"/ibc.applications.transfer.v1.MsgTransfer",
	}
	icaGenStateBz, err := cdc.MarshalJSON(icaGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal crisis genesis state: %w", err)
	}
	appState[icatypes.ModuleName] = icaGenStateBz

	// return appState and genDoc
	return appState, genDoc, nil
}
