package void

import (
	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	authzmodulev1 "cosmossdk.io/api/cosmos/authz/module/v1"
	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	circuitmodulev1 "cosmossdk.io/api/cosmos/circuit/module/v1"
	consensusmodulev1 "cosmossdk.io/api/cosmos/consensus/module/v1"
	distrmodulev1 "cosmossdk.io/api/cosmos/distribution/module/v1"
	evidencemodulev1 "cosmossdk.io/api/cosmos/evidence/module/v1"
	feegrantmodulev1 "cosmossdk.io/api/cosmos/feegrant/module/v1"
	genutilmodulev1 "cosmossdk.io/api/cosmos/genutil/module/v1"
	govmodulev1 "cosmossdk.io/api/cosmos/gov/module/v1"
	mintmodulev1 "cosmossdk.io/api/cosmos/mint/module/v1"
	paramsmodulev1 "cosmossdk.io/api/cosmos/params/module/v1"
	protocolpoolmodulev1 "cosmossdk.io/api/cosmos/protocolpool/module/v1"
	slashingmodulev1 "cosmossdk.io/api/cosmos/slashing/module/v1"
	stakingmodulev1 "cosmossdk.io/api/cosmos/staking/module/v1"
	txconfigv1 "cosmossdk.io/api/cosmos/tx/config/v1"
	upgrademodulev1 "cosmossdk.io/api/cosmos/upgrade/module/v1"
	vestingmodulev1 "cosmossdk.io/api/cosmos/vesting/module/v1"
	feemarketmodulev1 "github.com/skip-mev/feemarket/api/feemarket/feemarket/module/v1"

	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
	circuittypes "cosmossdk.io/x/circuit/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/skip-mev/feemarket/x/feemarket"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	_ "cosmossdk.io/x/circuit"                        // import for side-effects
	_ "cosmossdk.io/x/evidence"                       // import for side-effects
	_ "cosmossdk.io/x/feegrant/module"                // import for side-effects
	_ "cosmossdk.io/x/upgrade"                        // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/authz/module"   // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/bank"           // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/consensus"      // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/distribution"   // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/mint"           // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/protocolpool"   // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/slashing"       // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/staking"        // import for side-effects
)

var (
	// module account permissions
	moduleAccPerms = []*authmodulev1.ModuleAccountPermission{
		{Account: authtypes.FeeCollectorName},
		{Account: distrtypes.ModuleName},
		{Account: minttypes.ModuleName, Permissions: []string{authtypes.Minter}},
		{Account: stakingtypes.BondedPoolName, Permissions: []string{authtypes.Burner, stakingtypes.ModuleName}},
		{Account: stakingtypes.NotBondedPoolName, Permissions: []string{authtypes.Burner, stakingtypes.ModuleName}},
		{Account: govtypes.ModuleName, Permissions: []string{authtypes.Burner}},
		{Account: protocolpooltypes.ModuleName},
		{Account: protocolpooltypes.ProtocolPoolEscrowAccount},
		{Account: ibctransfertypes.ModuleName, Permissions: []string{authtypes.Minter, authtypes.Burner}},
		{Account: ibcfeetypes.ModuleName},
		{Account: icatypes.ModuleName},
		{Account: wasmtypes.ModuleName, Permissions: []string{authtypes.Burner}},
		{Account: feemarkettypes.ModuleName, Permissions: []string{authtypes.Burner}},
		{Account: feemarkettypes.FeeCollectorName, Permissions: []string{authtypes.Burner}},
	}

	// blocked account addresses
	blockAccAddrs = []string{
		authtypes.FeeCollectorName,
		distrtypes.ModuleName,
		minttypes.ModuleName,
		stakingtypes.BondedPoolName,
		stakingtypes.NotBondedPoolName,
		// We allow the following module accounts to receive funds:
		// govtypes.ModuleName
	}

	ModuleConfig = []*appv1alpha1.ModuleConfig{
		{
			Name: runtime.ModuleName,
			Config: appconfig.WrapAny(&runtimev1alpha1.Module{
				AppName: AppName,
				// NOTE: upgrade module is required to be prioritized
				PreBlockers: []string{
					upgradetypes.ModuleName,
					authtypes.ModuleName,
				},
				// During begin block slashing happens after distr.BeginBlocker so that
				// there is nothing left over in the validator fee pool, so as to keep the
				// CanWithdrawInvariant invariant.
				// NOTE: staking module is required if HistoricalEntries param > 0
				BeginBlockers: []string{
					minttypes.ModuleName,
					feemarkettypes.ModuleName,
					distrtypes.ModuleName,
					protocolpooltypes.ModuleName,
					slashingtypes.ModuleName,
					evidencetypes.ModuleName,
					stakingtypes.ModuleName,
					authz.ModuleName,
					// ibc modules
					capabilitytypes.ModuleName,
					ibcexported.ModuleName,
					ibctransfertypes.ModuleName,
					icatypes.ModuleName,
					ibcfeetypes.ModuleName,
					wasmtypes.ModuleName,
				},
				EndBlockers: []string{
					feemarkettypes.ModuleName,
					govtypes.ModuleName,
					stakingtypes.ModuleName,
					feegrant.ModuleName,
					protocolpooltypes.ModuleName,
					ibcexported.ModuleName,
					ibctransfertypes.ModuleName,
					capabilitytypes.ModuleName,
					icatypes.ModuleName,
					ibcfeetypes.ModuleName,
					wasmtypes.ModuleName,
				},
				OverrideStoreKeys: []*runtimev1alpha1.StoreKeyConfig{
					{
						ModuleName: authtypes.ModuleName,
						KvStoreKey: "acc",
					},
				},
				SkipStoreKeys: []string{
					"tx",
				},
				// NOTE: The genutils module must occur after staking so that pools are
				// properly initialized with tokens from genesis accounts.
				// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
				InitGenesis: []string{
					capabilitytypes.ModuleName,
					authtypes.ModuleName,
					banktypes.ModuleName,
					distrtypes.ModuleName,
					stakingtypes.ModuleName,
					slashingtypes.ModuleName,
					govtypes.ModuleName,
					minttypes.ModuleName,
					ibcexported.ModuleName,
					genutiltypes.ModuleName,
					evidencetypes.ModuleName,
					authz.ModuleName,
					ibctransfertypes.ModuleName,
					icatypes.ModuleName,
					ibcfeetypes.ModuleName,
					feegrant.ModuleName,
					paramstypes.ModuleName,
					upgradetypes.ModuleName,
					vestingtypes.ModuleName,
					circuittypes.ModuleName,
					protocolpooltypes.ModuleName,
					wasmtypes.ModuleName,
					feemarkettypes.ModuleName,
				},
				// When ExportGenesis is not specified, the export genesis module order
				// is equal to the init genesis order
				ExportGenesis: []string{
					capabilitytypes.ModuleName,
					authtypes.ModuleName,
					protocolpooltypes.ModuleName, // Must be exported before bank
					banktypes.ModuleName,
					distrtypes.ModuleName,
					stakingtypes.ModuleName,
					slashingtypes.ModuleName,
					govtypes.ModuleName,
					minttypes.ModuleName,
					ibcexported.ModuleName,
					genutiltypes.ModuleName,
					evidencetypes.ModuleName,
					authz.ModuleName,
					ibctransfertypes.ModuleName,
					icatypes.ModuleName,
					ibcfeetypes.ModuleName,
					feegrant.ModuleName,
					paramstypes.ModuleName,
					upgradetypes.ModuleName,
					vestingtypes.ModuleName,
					circuittypes.ModuleName,
					wasmtypes.ModuleName,
					feemarkettypes.ModuleName,
				},
				// Uncomment if you want to set a custom migration order here.
				// OrderMigrations: []string{},
			}),
		},
		{
			Name: authtypes.ModuleName,
			Config: appconfig.WrapAny(&authmodulev1.Module{
				Bech32Prefix:             AccountAddressPrefix,
				ModuleAccountPermissions: moduleAccPerms,
				// By default modules authority is the governance module. This is configurable with the following:
				// Authority: "group", // A custom module authority can be set using a module name
				// Authority: "vo1cwwv22j5ca08ggdv9c2uky355k908694z577tv", // or a specific address
			}),
		},
		{
			Name:   vestingtypes.ModuleName,
			Config: appconfig.WrapAny(&vestingmodulev1.Module{}),
		},
		{
			Name: banktypes.ModuleName,
			Config: appconfig.WrapAny(&bankmodulev1.Module{
				BlockedModuleAccountsOverride: blockAccAddrs,
			}),
		},
		{
			Name: stakingtypes.ModuleName,
			Config: appconfig.WrapAny(&stakingmodulev1.Module{
				// NOTE: specifying a prefix is only necessary when using bech32 addresses
				// If not specfied, the auth Bech32Prefix appended with "valoper" and "valcons" is used by default
				Bech32PrefixValidator: AccountAddressPrefix + "valoper",
				Bech32PrefixConsensus: AccountAddressPrefix + "valcons",
			}),
		},
		{
			Name:   slashingtypes.ModuleName,
			Config: appconfig.WrapAny(&slashingmodulev1.Module{}),
		},
		{
			Name: "tx",
			Config: appconfig.WrapAny(&txconfigv1.Config{
				SkipAnteHandler: true, // Enable this to skip the default antehandlers and set custom ante handlers.
			}),
		},
		{
			Name:   genutiltypes.ModuleName,
			Config: appconfig.WrapAny(&genutilmodulev1.Module{}),
		},
		{
			Name:   authz.ModuleName,
			Config: appconfig.WrapAny(&authzmodulev1.Module{}),
		},
		{
			Name:   upgradetypes.ModuleName,
			Config: appconfig.WrapAny(&upgrademodulev1.Module{}),
		},
		{
			Name:   distrtypes.ModuleName,
			Config: appconfig.WrapAny(&distrmodulev1.Module{}),
		},
		{
			Name:   evidencetypes.ModuleName,
			Config: appconfig.WrapAny(&evidencemodulev1.Module{}),
		},
		{
			Name:   minttypes.ModuleName,
			Config: appconfig.WrapAny(&mintmodulev1.Module{}),
		},
		{
			Name:   feegrant.ModuleName,
			Config: appconfig.WrapAny(&feegrantmodulev1.Module{}),
		},
		{
			Name:   govtypes.ModuleName,
			Config: appconfig.WrapAny(&govmodulev1.Module{}),
		},
		{
			Name:   paramstypes.ModuleName,
			Config: appconfig.WrapAny(&paramsmodulev1.Module{}),
		},
		{
			Name:   consensustypes.ModuleName,
			Config: appconfig.WrapAny(&consensusmodulev1.Module{}),
		},
		{
			Name:   circuittypes.ModuleName,
			Config: appconfig.WrapAny(&circuitmodulev1.Module{}),
		},
		{
			Name:   protocolpooltypes.ModuleName,
			Config: appconfig.WrapAny(&protocolpoolmodulev1.Module{}),
		},
		{
			Name:   feemarkettypes.ModuleName,
			Config: appconfig.WrapAny(&feemarketmodulev1.Module{}),
		},
	}

	// AppConfig is application configuration (used by depinject)
	AppConfig = depinject.Configs(appconfig.Compose(&appv1alpha1.Config{
		Modules: ModuleConfig,
	}),
		depinject.Supply(
			// supply custom module basics
			map[string]module.AppModuleBasic{
				genutiltypes.ModuleName:   genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
				govtypes.ModuleName:       gov.NewAppModuleBasic(getGovProposalHandlers()),
				feemarkettypes.ModuleName: feemarket.AppModuleBasic{},
			},
		),
	)
)
