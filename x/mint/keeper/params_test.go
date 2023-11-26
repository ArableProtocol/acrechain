package keeper_test

import (
	"time"

	"github.com/ArableProtocol/acrechain/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestParamsGetSet() {
	params := types.Params{
		MintDenom:                "aacre",
		GenesisDailyProvisions:   types.DefaultParams().GenesisDailyProvisions,
		ReductionPeriodInSeconds: 1000,
		ReductionFactor:          sdk.NewDecWithPrec(66, 2),
		DistributionProportions: types.DistributionProportions{
			Staking: sdk.NewDecWithPrec(2, 1),
		},
		NextRewardsReductionTime:            time.Now().Add(time.Second * 1000).Unix(),
		MintingRewardsDistributionStartTime: time.Now().Add(time.Second).Unix(),
	}

	suite.app.MintKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.MintKeeper.GetParams(suite.ctx)
	suite.Require().Equal(params, newParams)
}
