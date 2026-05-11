package keeper

import (
	"github.com/ArableProtocol/acrechain/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetNextReductionTime returns next reduction time.
func (k Keeper) GetNextReductionTime(ctx sdk.Context) int64 {
	params := k.GetParams(ctx)
	return params.NextRewardsReductionTime
}

// SetNextReductionTime set next reduction time.
func (k Keeper) SetNextReductionTime(ctx sdk.Context, time int64) {
	params := k.GetParams(ctx)
	params.NextRewardsReductionTime = time
	k.SetParams(ctx, params)
}

// get the minter.
func (k Keeper) GetMinter(ctx sdk.Context) (minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.MinterKey)
	if b == nil {
		panic("stored minter should not have been nil")
	}

	k.cdc.MustUnmarshal(b, &minter)
	return
}

// set the minter.
func (k Keeper) SetMinter(ctx sdk.Context, minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&minter)
	store.Set(types.MinterKey, b)
}
