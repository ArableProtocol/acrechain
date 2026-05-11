package keeper_test

import (
	"time"

	"github.com/ArableProtocol/acrechain/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestEndBlocker() {
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

	suite.SetupTest()
	suite.app.MintKeeper.SetParams(suite.ctx, params)

	now := time.Now()
	suite.ctx = suite.ctx.WithBlockTime(now)

	// check minter information at genesis
	minter := suite.app.MintKeeper.GetMinter(suite.ctx)
	suite.Require().Equal(minter.DailyProvisions, types.DefaultParams().GenesisDailyProvisions)
	suite.Require().Equal(minter.LastMintTime, int64(0))
	communityPool := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
	suite.Require().Equal(communityPool, sdk.DecCoins(nil))

	// run first endblocker
	suite.app.MintKeeper.EndBlocker(suite.ctx)

	// check everything is same
	minter = suite.app.MintKeeper.GetMinter(suite.ctx)
	suite.Require().Equal(minter.DailyProvisions, types.DefaultParams().GenesisDailyProvisions)
	suite.Require().Equal(minter.LastMintTime, int64(0))
	reductionTime := suite.app.MintKeeper.GetNextReductionTime(suite.ctx)
	suite.Require().Equal(reductionTime, params.NextRewardsReductionTime)
	communityPool = suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
	suite.Require().Equal(communityPool, sdk.DecCoins(nil))

	// run 2nd endblocker after a time
	suite.ctx = suite.ctx.WithBlockTime(now.Add(time.Second * 2))
	suite.app.MintKeeper.EndBlocker(suite.ctx)

	// check changes
	minter = suite.app.MintKeeper.GetMinter(suite.ctx)
	suite.Require().Equal(minter.DailyProvisions, types.DefaultParams().GenesisDailyProvisions)
	suite.Require().Equal(minter.LastMintTime, suite.ctx.BlockTime().Unix())
	reductionTime = suite.app.MintKeeper.GetNextReductionTime(suite.ctx)
	suite.Require().Equal(reductionTime, params.NextRewardsReductionTime)
	communityPool = suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
	suite.Require().Equal(communityPool, sdk.DecCoins(nil))

	// run 3rd endblocker after a time
	suite.ctx = suite.ctx.WithBlockTime(now.Add(time.Second * 3))
	suite.app.MintKeeper.EndBlocker(suite.ctx)

	// check changes
	minter = suite.app.MintKeeper.GetMinter(suite.ctx)
	suite.Require().Equal(minter.DailyProvisions, types.DefaultParams().GenesisDailyProvisions)
	suite.Require().Equal(minter.LastMintTime, suite.ctx.BlockTime().Unix())
	reductionTime = suite.app.MintKeeper.GetNextReductionTime(suite.ctx)
	suite.Require().Equal(reductionTime, params.NextRewardsReductionTime)
	communityPool = suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
	suite.Require().Equal(communityPool.String(), "7610342592592592592.000000000000000000aacre")

	// run 4th endblocker after reduction time
	suite.ctx = suite.ctx.WithBlockTime(now.Add(time.Second * 1001))
	suite.app.MintKeeper.EndBlocker(suite.ctx)

	// check changes
	minter = suite.app.MintKeeper.GetMinter(suite.ctx)
	suite.Require().Equal(minter.DailyProvisions, types.DefaultParams().GenesisDailyProvisions.Mul(params.ReductionFactor))
	suite.Require().Equal(minter.LastMintTime, suite.ctx.BlockTime().Unix())
	reductionTime = suite.app.MintKeeper.GetNextReductionTime(suite.ctx)
	suite.Require().Equal(reductionTime, suite.ctx.BlockTime().Unix()+params.ReductionPeriodInSeconds)
	params = suite.app.MintKeeper.GetParams(suite.ctx)
	suite.Require().Equal(params.NextRewardsReductionTime, reductionTime)
	communityPool = suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
	suite.Require().Equal(communityPool.String(), "5020390801481481481481.000000000000000000aacre")
}
