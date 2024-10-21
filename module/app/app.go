package app

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmos "github.com/cometbft/cometbft/libs/os"
	dbm "github.com/cosmos/cosmos-db"

	// Cosmos SDK

	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/evidence"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	sdkante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibcfeekeeper "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	ibctransfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	solomachine "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	// Cosmos IBC-Go
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	transfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v8/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	// Tharsis Ethermint
	ethante "github.com/evmos/ethermint/app/ante"

	"github.com/Gravity-Bridge/Gravity-Bridge/module/app/ante"
	gravityparams "github.com/Gravity-Bridge/Gravity-Bridge/module/app/params"
	gravityconfig "github.com/Gravity-Bridge/Gravity-Bridge/module/config"
	"github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity"
	"github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/keeper"
	gravitytypes "github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/keeper"

	"github.com/Gravity-Bridge/Gravity-Bridge/module/x/auction"
	auckeeper "github.com/Gravity-Bridge/Gravity-Bridge/module/x/auction/keeper"
	auctiontypes "github.com/Gravity-Bridge/Gravity-Bridge/module/x/auction/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
)

const appName = "app"

var (
	// DefaultNodeHome sets the folder where the applcation data and configuration will be stored
	DefaultNodeHome string

	// ModuleBasics The module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.

	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			[]govclient.ProposalHandler{
				paramsclient.ProposalHandler,
			},
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ibctm.AppModuleBasic{},
		solomachine.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
		ica.AppModuleBasic{},
		vesting.AppModuleBasic{},
		gravity.AppModuleBasic{},
		auction.AppModuleBasic{},
		ica.AppModuleBasic{},
		groupmodule.AppModuleBasic{},
	)

	// module account permissions
	// NOTE: We believe that this is giving various modules access to functions of the supply module? We will probably need to use this.
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:          nil,
		distrtypes.ModuleName:               nil,
		minttypes.ModuleName:                {authtypes.Minter},
		stakingtypes.BondedPoolName:         {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:      {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:                 {authtypes.Burner},
		ibctransfertypes.ModuleName:         {authtypes.Minter, authtypes.Burner},
		gravitytypes.ModuleName:             {authtypes.Minter, authtypes.Burner},
		auctiontypes.ModuleName:             {authtypes.Minter, authtypes.Burner},
		auctiontypes.AuctionPoolAccountName: nil,
		icatypes.ModuleName:                 nil,
	}

	// module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{
		distrtypes.ModuleName: true,
	}

	// verify app interface at compile time
	_ servertypes.Application = (*Gravity)(nil)
	_ runtime.AppI            = (*Gravity)(nil)

	// enable checks that run on the first BeginBlocker execution after an upgrade/genesis init/node restart
	firstBlock sync.Once
)

// Gravity extended ABCI application
type Gravity struct {
	*runtime.App
	legacyAmino       *codec.LegacyAmino
	AppCodec          codec.Codec
	txConfig          client.TxConfig
	InterfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tKeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	// NOTE: If you add anything to this struct, add a nil check to ValidateMembers below!
	AccountKeeper    *authkeeper.AccountKeeper
	AuthzKeeper      *authzkeeper.Keeper
	BankKeeper       *bankkeeper.BaseKeeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    *stakingkeeper.Keeper
	SlashingKeeper   *slashingkeeper.Keeper
	MintKeeper       *mintkeeper.Keeper
	DistrKeeper      *distrkeeper.Keeper
	GovKeeper        *govkeeper.Keeper
	CrisisKeeper     *crisiskeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper
	ParamsKeeper     *paramskeeper.Keeper
	IbcKeeper        *ibckeeper.Keeper
	EvidenceKeeper   *evidencekeeper.Keeper

	// IBC
	IBCKeeper           *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCFeeKeeper        ibcfeekeeper.Keeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	IbcTransferKeeper   ibctransferkeeper.Keeper

	// Scoped IBC
	// make scoped keepers public for test purposes
	// NOTE: If you add anything to this struct, add a nil check to ValidateMembers below!
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedIcaHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper

	GravityKeeper *keeper.Keeper
	AuctionKeeper *auckeeper.Keeper
	IcaHostKeeper *icahostkeeper.Keeper
	GroupKeeper   *groupkeeper.Keeper

	// simulation manager
	sm *module.SimulationManager
}

// MakeCodec creates the application codec. The codec is sealed before it is
// returned.
// func MakeCodec() *codec.LegacyAmino {
// 	var cdc = codec.NewLegacyAmino()
// 	ModuleBasics.RegisterLegacyAminoCodec(cdc)
// 	vesting.AppModuleBasic{}.RegisterLegacyAminoCodec(cdc)
// 	sdk.RegisterLegacyAminoCodec(cdc)
// 	ccodec.RegisterCrypto(cdc)
// 	cdc.Seal()
// 	return cdc
// }

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".gravity")
}

func NewGravityApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, skipUpgradeHeights map[int64]bool,
	homePath string, invCheckPeriod uint, encodingConfig gravityparams.EncodingConfig,
	appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp),
) *Gravity {
	var (
		app        = &Gravity{}
		appBuilder *runtime.AppBuilder
		appConfig  = depinject.Configs(
			depinject.Supply(
				// supply the application options
				appOpts,
				logger,
				// supply ibc keeper getter for the IBC modules
				app.GetIBCeKeeper,

				// supply the evm keeper getter for the EVM module
				// app.GetEVMKeeper,

				// ADVANCED CONFIGURATION
				//
				// AUTH
				//
				// For providing a custom function required in auth to generate custom account types
				// add it below. By default the auth module uses simulation.RandomGenesisAccounts.
				//
				// authtypes.RandomGenesisAccountsFn(simulation.RandomGenesisAccounts),

				// For providing a custom a base account type add it below.
				// By default the auth module uses authtypes.ProtoBaseAccount().
				//
				// func() authtypes.AccountI { return authtypes.ProtoBaseAccount() },

				//
				// MINT
				//

				// For providing a custom inflation function for x/mint add here your
				// custom function that implements the minttypes.InflationCalculationFn
				// interface.
			),
		)
	)
	if err := depinject.Inject(appConfig,
		&appBuilder,
		&app.AppCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.InterfaceRegistry,
		&app.AccountKeeper,
		&app.BankKeeper,
		&app.CapabilityKeeper,
		&app.StakingKeeper,
		&app.SlashingKeeper,
		&app.MintKeeper,
		&app.DistrKeeper,
		&app.GovKeeper,
		&app.CrisisKeeper,
		&app.UpgradeKeeper,
		&app.ParamsKeeper,
		&app.AuthzKeeper,
		&app.EvidenceKeeper,
		&app.GroupKeeper,
		&app.GravityKeeper,
		&app.AuctionKeeper,
		&app.IcaHostKeeper,
		&app.GroupKeeper,
		// IBC
		&app.IBCKeeper,
		&app.IBCFeeKeeper,
		&app.ICAControllerKeeper,
		&app.ICAHostKeeper,
		&app.IbcTransferKeeper,
		// this line is used by starport scaffolding # stargate/app/keeperDefinition
	); err != nil {
		panic(err)
	}

	encConfig := MakeEncodingConfig()
	app.AppCodec = encodingConfig.Marshaler
	app.legacyAmino = encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	app.SetTxEncoder(encConfig.TxConfig.TxEncoder())
	app.SetTxDecoder(encConfig.TxConfig.TxDecoder())

	initParamsKeeper(*app.ParamsKeeper)

	bApp := *baseapp.NewBaseApp(appName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, authzkeeper.StoreKey, banktypes.StoreKey,
		stakingtypes.StoreKey, minttypes.StoreKey, distrtypes.StoreKey,
		slashingtypes.StoreKey, govtypes.StoreKey, paramstypes.StoreKey,
		upgradetypes.StoreKey, evidencetypes.StoreKey,
		ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
		gravitytypes.StoreKey, auctiontypes.StoreKey,
		icahosttypes.StoreKey, group.StoreKey,
		ibcexported.StoreKey, ibctransfertypes.StoreKey,
		ibcfeetypes.StoreKey, icahosttypes.StoreKey,
		icacontrollertypes.StoreKey,
	)

	tKeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	// load state streaming if enabled
	if err := app.RegisterStreamingServices(appOpts, keys); err != nil {
		panic(err)
	}

	// nolint: exhaustruct
	// var app = &Gravity{
	// 	BaseApp:           &bApp,
	// 	legacyAmino:       legacyAmino,
	// 	AppCodec:          app.AppCodec,
	// 	InterfaceRegistry: interfaceRegistry,
	// 	invCheckPeriod:    invCheckPeriod,
	// 	keys:              keys,
	// 	tKeys:             tKeys,
	// 	memKeys:           memKeys,
	// }

	paramsKeeper := initParamsKeeper(*app.ParamsKeeper)
	app.ParamsKeeper = &paramsKeeper

	// paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
	// bApp.SetParamStore()

	capabilityKeeper := *capabilitykeeper.NewKeeper(
		app.AppCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)
	app.CapabilityKeeper = &capabilityKeeper

	scopedTransferKeeper := capabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	app.ScopedTransferKeeper = scopedTransferKeeper

	scopedIcaHostKeeper := capabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	app.ScopedIcaHostKeeper = scopedIcaHostKeeper

	// Applications that wish to enforce statically created ScopedKeepers should call `Seal` after creating
	// their scoped modules in `NewApp` with `ScopeToModule`
	capabilityKeeper.Seal()

	// get authority address
	authAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	accountKeeper := authkeeper.NewAccountKeeper(
		app.AppCodec,
		runtime.NewKVStoreService(app.GetKey(authtypes.StoreKey)),
		authtypes.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		// app.GetSubspace(authtypes.ModuleName),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authAddr,
	)
	app.AccountKeeper = &accountKeeper

	authzKeeper := authzkeeper.NewKeeper(
		runtime.NewKVStoreService(app.GetKey(authzkeeper.StoreKey)),
		app.AppCodec,
		app.MsgServiceRouter(),
		accountKeeper,
	)
	app.AuthzKeeper = &authzKeeper

	bankKeeper := bankkeeper.NewBaseKeeper(
		app.AppCodec,
		runtime.NewKVStoreService(app.GetKey(banktypes.StoreKey)),
		accountKeeper,
		app.BlockedAddrs(),
		authAddr,
		logger,
	)
	app.BankKeeper = &bankKeeper

	stakingKeeper := stakingkeeper.NewKeeper(
		app.AppCodec,
		runtime.NewKVStoreService(app.GetKey(stakingtypes.StoreKey)),
		accountKeeper,
		bankKeeper,
		authAddr,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		app.AccountKeeper.AddressCodec(),
	)
	app.StakingKeeper = stakingKeeper

	distrKeeper := distrkeeper.NewKeeper(
		app.AppCodec,
		runtime.NewKVStoreService(app.GetKey(distrtypes.StoreKey)),
		app.AccountKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authAddr,
	)
	app.DistrKeeper = &distrKeeper

	slashingKeeper := slashingkeeper.NewKeeper(
		app.AppCodec,
		app.legacyAmino,
		runtime.NewKVStoreService(app.GetKey(slashingtypes.StoreKey)),
		stakingKeeper,
		authAddr,
	)
	app.SlashingKeeper = &slashingKeeper

	upgradeKeeper := upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(app.GetKey(upgradetypes.StoreKey)),
		app.AppCodec,
		homePath,
		&bApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.UpgradeKeeper = upgradeKeeper

	app.ScopedIBCKeeper = app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	app.ScopedIcaHostKeeper = app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	app.ScopedICAControllerKeeper = app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)

	// ScopedIBCKeeper           *capabilitykeeper.ScopedKeeper
	// ScopedTransferKeeper      *capabilitykeeper.ScopedKeeper
	// ScopedIcaHostKeeper       *capabilitykeeper.ScopedKeeper
	// ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper

	ibcKeeper := *ibckeeper.NewKeeper(
		app.AppCodec,
		app.GetKey(ibcexported.StoreKey),
		app.GetSubspace(ibcexported.ModuleName),
		stakingKeeper,
		upgradeKeeper,
		app.ScopedIBCKeeper,
		authAddr,
	)
	app.IbcKeeper = &ibcKeeper

	ibcTransferKeeper := ibctransferkeeper.NewKeeper(
		app.AppCodec, keys[ibctransfertypes.StoreKey], app.GetSubspace(ibctransfertypes.ModuleName),
		ibcKeeper.ChannelKeeper, ibcKeeper.ChannelKeeper, ibcKeeper.PortKeeper,
		accountKeeper, bankKeeper, scopedTransferKeeper,
		authAddr,
	)
	app.IbcTransferKeeper = ibcTransferKeeper

	icaHostKeeper := icahostkeeper.NewKeeper(
		app.AppCodec, keys[icahosttypes.StoreKey], app.GetSubspace(icahosttypes.SubModuleName),
		ibcKeeper.ChannelKeeper, ibcKeeper.ChannelKeeper, ibcKeeper.PortKeeper,
		accountKeeper, scopedIcaHostKeeper, app.MsgServiceRouter(),
		authAddr,
	)
	app.IcaHostKeeper = &icaHostKeeper

	mintKeeper := mintkeeper.NewKeeper(
		app.AppCodec,
		runtime.NewKVStoreService(app.GetKey(minttypes.StoreKey)),
		stakingKeeper,
		accountKeeper,
		bankKeeper,
		authtypes.FeeCollectorName,
		authAddr,
	)
	app.MintKeeper = &mintKeeper

	auctionKeeper := auckeeper.NewKeeper(
		keys[auctiontypes.StoreKey],
		app.GetSubspace(auctiontypes.ModuleName),
		app.AppCodec,
		&bankKeeper,
		&accountKeeper,
		&distrKeeper,
		&mintKeeper,
	)
	app.AuctionKeeper = &auctionKeeper

	gravityKeeper := keeper.NewKeeper(
		keys[gravitytypes.StoreKey],
		app.GetSubspace(gravitytypes.ModuleName),
		app.AppCodec,
		&bankKeeper,
		stakingKeeper,
		&slashingKeeper,
		&distrKeeper,
		&accountKeeper,
		&ibcTransferKeeper,
		&auctionKeeper,
	)
	app.GravityKeeper = &gravityKeeper

	// Add the staking hooks from distribution, slashing, and gravity to staking
	stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			distrKeeper.Hooks(),
			slashingKeeper.Hooks(),
			gravityKeeper.Hooks(),
		),
	)

	crisisKeeper := crisiskeeper.NewKeeper(
		app.AppCodec,
		runtime.NewKVStoreService(app.GetKey(crisistypes.StoreKey)),
		invCheckPeriod,
		bankKeeper,
		authtypes.FeeCollectorName,
		authAddr,
		app.AccountKeeper.AddressCodec(),
	)
	app.CrisisKeeper = crisisKeeper

	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramsproposal.RouterKey, params.NewParamChangeProposalHandler(paramsKeeper)).
		// AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(distrKeeper)).
		// AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(upgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(ibcKeeper.ClientKeeper)).
		AddRoute(gravitytypes.RouterKey, keeper.NewGravityProposalHandler(gravityKeeper))

	govConfig := govtypes.DefaultConfig()
	govKeeper := govkeeper.NewKeeper(
		app.AppCodec,
		runtime.NewKVStoreService(app.GetKey(govtypes.StoreKey)),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		distrKeeper,
		app.MsgServiceRouter(),
		govConfig,
		authAddr,
	)
	govKeeper = govKeeper.SetHooks(govtypes.NewMultiGovHooks(
	// Register any governance hooks here
	))
	app.GovKeeper = govKeeper

	// ibcTransferAppModule := transfer.NewAppModule(ibcTransferKeeper)
	ibcTransferIBCModule := transfer.NewIBCModule(ibcTransferKeeper)
	// icaAppModule := ica.NewAppModule(nil, &icaHostKeeper)
	icaHostIBCModule := icahost.NewIBCModule(icaHostKeeper)

	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, ibcTransferIBCModule).
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule)
	ibcKeeper.SetRouter(ibcRouter)

	// evidenceKeeper := *evidencekeeper.NewKeeper(
	// 	app.AppCodec,
	// 	keys[evidencetypes.StoreKey],
	// 	&stakingKeeper,
	// 	slashingKeeper,
	// 	app.AccountKeeper.AddressCodec(),
	// )
	// app.EvidenceKeeper = &evidenceKeeper

	groupConfig := group.DefaultConfig()
	groupKeeper := groupkeeper.NewKeeper(keys[group.StoreKey], app.AppCodec, app.MsgServiceRouter(), app.AccountKeeper, groupConfig)
	app.GroupKeeper = &groupKeeper

	// NOTE: capability module's BeginBlocker must come before any modules using capabilities (e.g. IBC)

	// sm := *module.NewSimulationManager(
	// 	auth.NewAppModule(app.AppCodec, accountKeeper, authsims.RandomGenesisAccounts),
	// 	bank.NewAppModule(app.AppCodec, bankKeeper, accountKeeper),
	// 	capability.NewAppModule(app.AppCodec, capabilityKeeper),
	// 	gov.NewAppModule(app.AppCodec, govKeeper, accountKeeper, bankKeeper),
	// 	mint.NewAppModule(app.AppCodec, mintKeeper, accountKeeper, nil),
	// 	staking.NewAppModule(app.AppCodec, stakingKeeper, accountKeeper, bankKeeper),
	// 	distr.NewAppModule(app.AppCodec, distrKeeper, accountKeeper, bankKeeper, stakingKeeper),
	// 	slashing.NewAppModule(app.AppCodec, slashingKeeper, accountKeeper, bankKeeper, stakingKeeper),
	// 	params.NewAppModule(paramsKeeper),
	// 	evidence.NewAppModule(evidenceKeeper),
	// 	ibc.NewAppModule(&ibcKeeper),
	// 	ibcTransferAppModule,
	// )
	// app.sm = &sm

	// sm.RegisterStoreDecoders()

	app.MountKVStores(keys)
	app.MountTransientStores(tKeys)
	app.MountMemoryStores(memKeys)

	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)

	app.SetEndBlocker(app.EndBlocker)

	app.setAnteHandler(encodingConfig)
	app.setPostHandler()

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	keeper.RegisterProposalTypes()

	// We don't allow anything to be nil
	return app
}

func (app *Gravity) setAnteHandler(encodingConfig gravityparams.EncodingConfig) {
	options := sdkante.HandlerOptions{
		AccountKeeper:          app.AccountKeeper,
		BankKeeper:             app.BankKeeper,
		FeegrantKeeper:         nil,
		SignModeHandler:        encodingConfig.TxConfig.SignModeHandler(),
		SigGasConsumer:         ethante.DefaultSigVerificationGasConsumer,
		ExtensionOptionChecker: nil,
		TxFeeChecker:           nil,
	}

	// Note: If feegrant keeper is added, add it to the NewAnteHandler call instead of nil
	ah, err := ante.NewAnteHandler(options, app.GravityKeeper, app.AccountKeeper, app.BankKeeper, nil, app.IbcKeeper, app.AppCodec, gravityconfig.GravityEvmChainID)
	if err != nil {
		panic("invalid antehandler created")
	}
	app.SetAnteHandler(*ah)
}
func (app *Gravity) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

func MakeCodecs() (codec.Codec, *codec.LegacyAmino) {
	config := MakeEncodingConfig()
	return config.Marshaler, config.Amino
}

// Name returns the name of the App
func (app *Gravity) Name() string { return app.BaseApp.Name() }

// Perform necessary checks at the start of this node's first BeginBlocker execution
// Note: This should ONLY be called once, it should be called at the top of BeginBlocker guarded by firstBlock
func (app *Gravity) firstBeginBlocker(ctx sdk.Context) {
	app.assertNativeTokenMatchesConstant(ctx)
	app.assertNativeTokenIsNonAuctionable(ctx)
}

// InitChainer application update at chain initialization
func (app *Gravity) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())

	return app.ModuleManager.InitGenesis(ctx, app.AppCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *Gravity) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *Gravity) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlockedAddrs returns all the app's module account addresses that are not
// allowed to receive external tokens.
func (app *Gravity) BlockedAddrs() map[string]bool {
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	}

	return blockedAddrs
}

// GetSubspace returns a param subspace for a given module name.
func (app *Gravity) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface
func (app *Gravity) SimulationManager() *module.SimulationManager {
	return app.sm
}
func (app *Gravity) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// GetTxConfig implements the TestingApp interface.
func (app *Gravity) GetTxConfig() client.TxConfig {
	cfg := MakeEncodingConfig()
	return cfg.TxConfig
}

// GetBaseApp returns the base app of the application
func (app *Gravity) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *Gravity) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	if apiConfig.Swagger {
		RegisterSwaggerAPI(clientCtx, apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *Gravity) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.InterfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *Gravity) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.InterfaceRegistry, app.Query)
}

func (app *Gravity) RegisterNodeService(clientCtx client.Context, conf config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), conf)
}

// RegisterSwaggerAPI registers swagger route with API Server
// TODO: build the custom gravity swagger files and add here?
func RegisterSwaggerAPI(ctx client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// GetMaccPerms returns a mapping of the application's module account permissions.
func GetMaccPerms() map[string][]string {
	modAccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		modAccPerms[k] = v
	}
	return modAccPerms
}

// GetIBCeKeeper returns the IBC keeper
func (app *Gravity) GetIBCeKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetKey returns the KVStoreKey for the provided store key.
func (app *Gravity) GetKey(storeKey string) *storetypes.KVStoreKey {
	if key, ok := app.keys[storeKey]; ok {
		return key
	}

	sk := app.UnsafeFindStoreKey(storeKey)
	kvStoreKey, ok := sk.(*storetypes.KVStoreKey)
	if !ok {
		return nil
	}
	return kvStoreKey
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(paramsKeeper paramskeeper.Keeper) paramskeeper.Keeper {

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(gravitytypes.ModuleName)
	paramsKeeper.Subspace(auctiontypes.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)

	return paramsKeeper
}
