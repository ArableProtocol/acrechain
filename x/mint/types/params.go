package types

import (
	"errors"
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/ArableProtocol/acrechain/cmd/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyMintDenom                           = []byte("MintDenom")
	KeyGenesisDailyProvisions              = []byte("GenesisDailyProvisions")
	KeyReductionPeriodInSeconds            = []byte("ReductionPeriodInSeconds")
	KeyReductionFactor                     = []byte("ReductionFactor")
	KeyDistributionProportions             = []byte("DistributionProportions")
	KeyMintingRewardsDistributionStartTime = []byte("MintingRewardsDistributionStartTime")
	KeyNextRewardsReductionTime            = []byte("NextRewardsReductionTime")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams returns new mint module parameters initialized to the given values.
func NewParams(
	mintDenom string, genesisDailyProvisions sdk.Dec,
	ReductionFactor sdk.Dec, reductionPeriodInSeconds int64, distrProportions DistributionProportions,
	nextRewardsReductionTime int64,
	mintingRewardsDistributionStartTime int64,
) Params {
	return Params{
		MintDenom:                           mintDenom,
		GenesisDailyProvisions:              genesisDailyProvisions,
		ReductionPeriodInSeconds:            reductionPeriodInSeconds,
		ReductionFactor:                     ReductionFactor,
		DistributionProportions:             distrProportions,
		NextRewardsReductionTime:            nextRewardsReductionTime,
		MintingRewardsDistributionStartTime: mintingRewardsDistributionStartTime,
	}
}

// DefaultParams returns the default minting module parameters.
func DefaultParams() Params {
	reductionDec := sdk.NewDecFromInt(config.DecimalReduction)
	return Params{
		MintDenom:                sdk.DefaultBondDenom,
		GenesisDailyProvisions:   sdk.NewDec(821_917).Mul(reductionDec), //  300 million /  365 * 10^18
		ReductionPeriodInSeconds: 31536000,                              // 1 year - 86400 x 365
		ReductionFactor:          sdk.NewDecWithPrec(6666, 4),           // 0.6666
		DistributionProportions: DistributionProportions{
			Staking: sdk.NewDecWithPrec(25, 2), // 25%
		},
		NextRewardsReductionTime:            0,
		MintingRewardsDistributionStartTime: 0,
	}
}

// Validate validates mint module parameters. Returns nil if valid,
// error otherwise
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateGenesisDailyProvisions(p.GenesisDailyProvisions); err != nil {
		return err
	}
	if err := validateReductionPeriodInSeconds(p.ReductionPeriodInSeconds); err != nil {
		return err
	}
	if err := validateReductionFactor(p.ReductionFactor); err != nil {
		return err
	}
	if err := validateDistributionProportions(p.DistributionProportions); err != nil {
		return err
	}

	if err := validateNextRewardsReductionTime(p.NextRewardsReductionTime); err != nil {
		return err
	}

	if err := validateMintingRewardsDistributionStartTime(p.MintingRewardsDistributionStartTime); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {

	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeyGenesisDailyProvisions, &p.GenesisDailyProvisions, validateGenesisDailyProvisions),
		paramtypes.NewParamSetPair(KeyReductionPeriodInSeconds, &p.ReductionPeriodInSeconds, validateReductionPeriodInSeconds),
		paramtypes.NewParamSetPair(KeyReductionFactor, &p.ReductionFactor, validateReductionFactor),
		paramtypes.NewParamSetPair(KeyDistributionProportions, &p.DistributionProportions, validateDistributionProportions),
		paramtypes.NewParamSetPair(KeyNextRewardsReductionTime, &p.NextRewardsReductionTime, validateNextRewardsReductionTime),
		paramtypes.NewParamSetPair(KeyMintingRewardsDistributionStartTime, &p.MintingRewardsDistributionStartTime, validateMintingRewardsDistributionStartTime),
	}
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateGenesisDailyProvisions(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(sdk.ZeroDec()) {
		return fmt.Errorf("genesis block provision must be non-negative")
	}

	return nil
}

func validateReductionPeriodInSeconds(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("reduction period must be positive: %d", v)
	}

	return nil
}

func validateReductionFactor(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.GT(sdk.NewDec(1)) {
		return fmt.Errorf("reduction factor cannot be greater than 1")
	}

	if v.IsNegative() {
		return fmt.Errorf("reduction factor cannot be negative")
	}

	return nil
}

func validateDistributionProportions(i interface{}) error {
	v, ok := i.(DistributionProportions)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Staking.IsNegative() {
		return errors.New("staking distribution ratio should not be negative")
	}

	return nil
}

func validateNextRewardsReductionTime(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 0 {
		return fmt.Errorf("next reduction time must be non-negative")
	}

	return nil
}

func validateMintingRewardsDistributionStartTime(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 0 {
		return fmt.Errorf("start time must be non-negative")
	}

	return nil
}
