package app

import (
	"fmt"

	"github.com/Gravity-Bridge/Gravity-Bridge/module/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// The community pool holds a significant balance of GRAV, so to make sure it cannot be auctioned off
// (which would have to be for MUCH less GRAV than it is worth), assert that the NonAuctionableTokens list
// contains GRAV (ugraviton)
func (app *Gravity) assertNativeTokenIsNonAuctionable(ctx sdk.Context) {
	nonAuctionableTokens := app.AuctionKeeper.GetParams(ctx).NonAuctionableTokens
	params, _ := app.MintKeeper.Params.Get(ctx)
	nativeToken := params.MintDenom // GRAV

	for _, t := range nonAuctionableTokens {
		if t == nativeToken {
			// Success!
			return
		}
	}

	// Failure!
	panic(fmt.Sprintf("Auction module's nonAuctionableTokens (%v) MUST contain GRAV (%s)\n", nonAuctionableTokens, nativeToken))
}

// In the config directory is a constant which should represent the native token, this check ensures that constant is correct
func (app *Gravity) assertNativeTokenMatchesConstant(ctx sdk.Context) {
	hardcoded := config.NativeTokenDenom
	params, _ := app.MintKeeper.Params.Get(ctx)
	nativeToken := params.MintDenom

	if hardcoded != nativeToken {
		panic(fmt.Sprintf("The hard-coded native token denom (%s) must equal the actual native token (%s)\n", hardcoded, nativeToken))
	}
}
