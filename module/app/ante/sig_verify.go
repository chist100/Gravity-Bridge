package ante

import (
	errorsmod "cosmossdk.io/errors"

	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	ethermintante "github.com/evmos/ethermint/app/ante"
)

// Gravity supports both ordinary SDK signing (legacy amino and sign direct), but also has EIP712 support.
// The Ethermint ante handler supports EVM messages and also maintains multiple lists of ante decorators,
// which is undesireable for Gravity. Instead this decorator will delegate the signing to the correct
// function depending on the type of transaction being processed.
//
// More accurately, if the input transaction has EXACTLY ONE ExtensionOptionsWeb3Tx option on it, we
// delegate to ethermint's Eip712SigVerificationDecorator.
// Otherwise we delegate to the SDK's SigVerificationDecorator.
type GravitySigVerificationDecorator struct {
	cdc    codec.Codec
	sdkSVD sdkante.SigVerificationDecorator
	// nolint: staticcheck
	ethermintSVD ethermintante.LegacyEip712SigVerificationDecorator
}

// See GravitySigVerificationDecorator for more info
func NewGravitySigVerificationDecorator(
	cdc codec.Codec, ak *authkeeper.AccountKeeper, signModeHandler txsigning.HandlerMap, evmChainID string,
) GravitySigVerificationDecorator {
	sdkSVD := sdkante.NewSigVerificationDecorator(ak, &signModeHandler)
	// nolint: staticcheck
	ethermintSVD := ethermintante.NewLegacyEip712SigVerificationDecorator(ak, &signModeHandler)
	return GravitySigVerificationDecorator{cdc, sdkSVD, ethermintSVD}
}

// See GravitySigVerificationDecorator for more info
func (svd GravitySigVerificationDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	isEIP712Signed, err := IsWeb3Tx(svd.cdc, tx)
	if err != nil {
		return ctx, errorsmod.Wrap(err, "unexpected tx extension options")
	}

	if isEIP712Signed {
		return svd.ethermintSVD.AnteHandle(ctx, tx, simulate, next)
	} else {
		return svd.sdkSVD.AnteHandle(ctx, tx, simulate, next)
	}
}
