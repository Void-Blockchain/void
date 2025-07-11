package void

import (
	"fmt"

	"cosmossdk.io/core/appmodule"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	pfm "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward"
	pfmkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/keeper"
	pfmtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/types"
	ratelimit "github.com/cosmos/ibc-apps/modules/rate-limiting/v10"
	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/keeper"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
	ratelimitv2 "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/v2"
	ica "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	ibccallbacks "github.com/cosmos/ibc-go/v10/modules/apps/callbacks"
	ibccallbacksv2 "github.com/cosmos/ibc-go/v10/modules/apps/callbacks/v2"
	ibctransfer "github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibctransferv2 "github.com/cosmos/ibc-go/v10/modules/apps/transfer/v2"
	ibc "github.com/cosmos/ibc-go/v10/modules/core"
	ibcclienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcapi "github.com/cosmos/ibc-go/v10/modules/core/api"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"
	solomachine "github.com/cosmos/ibc-go/v10/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

// RegisterLegacyModules register legacy keepers and non dependency inject modules.
func (app *VoidApp) RegisterLegacyModules(appOpts servertypes.AppOptions) error {
	// set up non depinject support modules store keys
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(ibcexported.StoreKey),
		storetypes.NewKVStoreKey(ibctransfertypes.StoreKey),
		storetypes.NewKVStoreKey(icahosttypes.StoreKey),
		storetypes.NewKVStoreKey(icacontrollertypes.StoreKey),
		storetypes.NewKVStoreKey(pfmtypes.StoreKey),
		storetypes.NewKVStoreKey(ratelimittypes.StoreKey),
		storetypes.NewKVStoreKey(wasmtypes.StoreKey),
	); err != nil {
		return err
	}

	// register the key tables for legacy param subspaces
	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	app.ParamsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(keyTable)
	app.ParamsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())
	app.ParamsKeeper.Subspace(icacontrollertypes.SubModuleName).WithKeyTable(icacontrollertypes.ParamKeyTable())
	app.ParamsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())
	app.ParamsKeeper.Subspace(pfmtypes.ModuleName)
	app.ParamsKeeper.Subspace(ratelimittypes.ModuleName).WithKeyTable(ratelimittypes.ParamKeyTable())

	govModuleAddr, _ := app.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))

	// Create IBC keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(ibcexported.StoreKey)),
		app.GetSubspace(ibcexported.ModuleName),
		app.UpgradeKeeper,
		govModuleAddr,
	)

	// Create interchain account keepers
	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(icahosttypes.StoreKey)),
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // ICS4Wrapper
		app.IBCKeeper.ChannelKeeper,
		app.AccountKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		govModuleAddr,
	)

	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(icacontrollertypes.StoreKey)),
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.MsgServiceRouter(),
		govModuleAddr,
	)

	// Create RateLimit keeper
	app.RateLimitKeeper = *ratelimitkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(ratelimittypes.StoreKey)),
		app.GetSubspace(ratelimittypes.ModuleName),
		govModuleAddr,
		app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ClientKeeper,
		app.IBCKeeper.ChannelKeeper,
	)

	// Packet Forward Middleware keeper
	// PFMKeeper must be created before TransferKeeper
	app.PFMKeeper = pfmkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(pfmtypes.StoreKey)),
		nil,
		app.IBCKeeper.ChannelKeeper,
		app.BankKeeper,
		app.RateLimitKeeper, // ICS4Wrapper
		govModuleAddr,
	)

	// Create IBC transfer keeper
	app.IBCTransferKeeper = ibctransferkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(ibctransfertypes.StoreKey)),
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		app.BankKeeper,
		govModuleAddr,
	)

	// Must be called on PFMRouter AFTER TransferKeeper initialized
	app.PFMKeeper.SetTransferKeeper(app.IBCTransferKeeper)

	// Wasm Module
	wasmConfig, err := wasm.ReadNodeConfig(appOpts)
	if err != nil {
		return fmt.Errorf("error while reading wasm config: %s", err)
	}

	wasmOpts := []wasmkeeper.Option{}
	app.WasmKeeper = wasmkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(wasmtypes.StoreKey)),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeperV2,
		app.IBCTransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		DefaultNodeHome,
		wasmConfig,
		wasmtypes.VMConfig{},
		wasmkeeper.BuiltInCapabilities(),
		govModuleAddr,
		wasmOpts...,
	)

	var (
		wasmStackIBCHandler wasm.IBCHandler
		icaControllerStack  porttypes.IBCModule
		noAuthzModule       porttypes.IBCModule
		icaHostStack        porttypes.IBCModule
		transferStack       porttypes.IBCModule
	)

	// create wasm stack
	wasmStackIBCHandler = wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCTransferKeeper, app.IBCKeeper.ChannelKeeper)

	// Create Transfer Stack
	transferStack = ibctransfer.NewIBCModule(app.IBCTransferKeeper)
	cbStack := ibccallbacks.NewIBCMiddleware(transferStack, app.PFMKeeper, wasmStackIBCHandler, wasm.DefaultMaxIBCCallbackGas)
	transferStack = pfm.NewIBCMiddleware(
		cbStack,
		app.PFMKeeper,
		0, // retries on timeout
		pfmkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
	)
	transferStack = ratelimit.NewIBCMiddleware(app.RateLimitKeeper, transferStack)
	var transferICS4Wrapper porttypes.ICS4Wrapper = cbStack
	// Since the callbacks middleware itself is an ics4wrapper, it needs to be passed to the ica controller keeper
	app.IBCTransferKeeper.WithICS4Wrapper(transferICS4Wrapper)

	// Create Interchain Accounts Stack
	icaControllerStack = icacontroller.NewIBCMiddlewareWithAuth(noAuthzModule, app.ICAControllerKeeper)
	icaControllerStack = icacontroller.NewIBCMiddlewareWithAuth(icaControllerStack, app.ICAControllerKeeper)
	icaControllerStack = ibccallbacks.NewIBCMiddleware(icaControllerStack, app.IBCKeeper.ChannelKeeper, wasmStackIBCHandler, wasm.DefaultMaxIBCCallbackGas)
	icaICS4Wrapper := icaControllerStack.(porttypes.ICS4Wrapper)
	// Since the callbacks middleware itself is an ics4wrapper, it needs to be passed to the ica controller keeper
	app.ICAControllerKeeper.WithICS4Wrapper(icaICS4Wrapper)

	// RecvPacket, message that originates from core IBC and goes down to app, the flow is:
	// channel.RecvPacket -> icaHost.OnRecvPacket
	icaHostStack = icahost.NewIBCModule(app.ICAHostKeeper)

	// set denom resolver to test variant.
	app.FeeMarketKeeper.SetDenomResolver(&feemarkettypes.TestDenomResolver{})

	// Create static IBC router
	ibcRouter := porttypes.NewRouter().
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icahosttypes.SubModuleName, icaHostStack).
		AddRoute(wasmtypes.ModuleName, wasmStackIBCHandler)

	app.IBCKeeper.SetRouter(ibcRouter)

	// IBCv2 transfer stack
	var transferStackV2 ibcapi.IBCModule
	transferStackV2 = ibctransferv2.NewIBCModule(app.IBCTransferKeeper)
	transferStackV2 = ibccallbacksv2.NewIBCMiddleware(transferStackV2, app.IBCKeeper.ChannelKeeperV2,
		wasmStackIBCHandler, app.IBCKeeper.ChannelKeeperV2, wasm.DefaultMaxIBCCallbackGas)
	transferStackV2 = ratelimitv2.NewIBCMiddleware(app.RateLimitKeeper, transferStackV2)

	// Create IBCv2 Router & seal
	ibcRouterV2 := ibcapi.NewRouter()
	ibcRouterV2 = ibcRouterV2.
		AddRoute(ibctransfertypes.PortID, transferStackV2).
		AddPrefixRoute(wasmkeeper.PortIDPrefixV2, wasmkeeper.NewIBC2Handler(app.WasmKeeper))
	app.IBCKeeper.SetRouterV2(ibcRouterV2)

	clientKeeper := app.IBCKeeper.ClientKeeper
	storeProvider := clientKeeper.GetStoreProvider()

	tmLightClientModule := ibctm.NewLightClientModule(app.appCodec, storeProvider)
	clientKeeper.AddRoute(ibctm.ModuleName, &tmLightClientModule)

	soloLightClientModule := solomachine.NewLightClientModule(app.appCodec, storeProvider)
	clientKeeper.AddRoute(solomachine.ModuleName, &soloLightClientModule)

	// register legacy modules
	if err := app.RegisterModules(
		// WASM modules
		wasm.NewAppModule(
			app.AppCodec(),
			&app.WasmKeeper,
			app.StakingKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.MsgServiceRouter(),
			app.GetSubspace(wasmtypes.ModuleName),
		),
		// IBC modules
		ibc.NewAppModule(app.IBCKeeper),
		ibctransfer.NewAppModule(app.IBCTransferKeeper),
		ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper),
		pfm.NewAppModule(app.PFMKeeper, app.GetSubspace(pfmtypes.ModuleName)),
		ratelimit.NewAppModule(app.appCodec, app.RateLimitKeeper),

		// IBC lightclient
		ibctm.NewAppModule(tmLightClientModule),
		solomachine.NewAppModule(soloLightClientModule),
	); err != nil {
		return err
	}

	// ante
	if err := app.setAnteHandler(app.txConfig, wasmConfig, app.GetKey(wasmtypes.StoreKey)); err != nil {
		return err
	}

	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
		)
		if err != nil {
			return fmt.Errorf("failed to register snapshot extension: %s", err)
		}
	}

	// post
	if err := app.setPostHandler(); err != nil {
		return err
	}

	return nil
}

// RegisterLegacyCLI Since the some modules don't support dependency injection,
// we need to manually register the modules on the client side.
func RegisterLegacyCLI(cdc codec.Codec) map[string]appmodule.AppModule {
	modules := map[string]appmodule.AppModule{
		// ibc
		ibcexported.ModuleName:      ibc.AppModule{},
		ibctransfertypes.ModuleName: ibctransfer.AppModule{},
		icatypes.ModuleName:         ica.AppModule{},
		pfmtypes.ModuleName:         pfm.AppModule{},
		ratelimittypes.ModuleName:   ratelimit.AppModule{},
		// lightclient
		ibctm.ModuleName:       ibctm.AppModule{},
		solomachine.ModuleName: solomachine.AppModule{},
		// wasm
		wasmtypes.ModuleName: wasm.AppModule{},
	}

	for _, m := range modules {
		if mr, ok := m.(module.AppModuleBasic); ok {
			mr.RegisterInterfaces(cdc.InterfaceRegistry())
		}
	}

	return modules
}

// GetSubspace returns a param subspace for a given module name.
func (app *VoidApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// GetMemKey returns the MemoryStoreKey for the provided store key.
func (app *VoidApp) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	key, ok := app.UnsafeFindStoreKey(storeKey).(*storetypes.MemoryStoreKey)
	if !ok {
		return nil
	}

	return key
}

func (app *VoidApp) setAnteHandler(txConfig client.TxConfig, wasmConfig wasmtypes.NodeConfig, txCounterStoreKey *storetypes.KVStoreKey) error {
	anteHandler, err := NewAnteHandler(
		AnteHandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				SignModeHandler: txConfig.SignModeHandler(),
				FeegrantKeeper:  app.FeeGrantKeeper,
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},

			IBCKeeper:             app.IBCKeeper,
			WasmConfig:            &wasmConfig,
			WasmKeeper:            &app.WasmKeeper,
			TXCounterStoreService: runtime.NewKVStoreService(txCounterStoreKey),
			CircuitKeeper:         &app.CircuitKeeper,
			AccountKeeper:         &app.AccountKeeper,
			BankKeeper:            app.BankKeeper,
			FeeMarketKeeper:       &app.FeeMarketKeeper,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create AnteHandler: %s", err)
	}

	// Set the AnteHandler for the app
	app.SetAnteHandler(anteHandler)
	return nil
}

func (app *VoidApp) setPostHandler() error {
	postHandlerOptions := PostHandlerOptions{
		AccountKeeper:   app.AccountKeeper,
		BankKeeper:      app.BankKeeper,
		FeeMarketKeeper: &app.FeeMarketKeeper,
	}

	postHandler, err := NewPostHandler(postHandlerOptions)
	if err != nil {
		return err
	}

	app.SetPostHandler(postHandler)
	return nil
}
