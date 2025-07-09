#!/usr/bin/env bash
set -euo pipefail

BINARY="void"           
CHAIN_HOME="$HOME/.void"   
CHAIN_ID="void-tesnet"
DENOM="uvoid"

echo "🔄 Rebuild binary..."
make install

echo "🧹 Reset data..."
rm -rf "$CHAIN_HOME"

echo "🔧 Initialize chain..."
$BINARY init validator --home "$CHAIN_HOME" --chain-id "$CHAIN_ID"

# Change default denom on genesis.json from 'stake' to $DENOM
echo "🧪 Change denom stake → $DENOM in genesis..."
jq --arg d "$DENOM" '
  (.app_state.feemarket.params.fee_denom) = $d |
  (.app_state.staking.params.bond_denom) = $d |
  (.app_state.mint.params.mint_denom) = $d |
  (.app_state.gov.params.min_deposit[] .denom) = $d |
  (.app_state.gov.params.expedited_min_deposit[] .denom) = $d |
  (.app_state.protocolpool.params.enabled_distribution_denoms[]
     ) |= (if . == "stake" then $d else . end)
' "$CHAIN_HOME/config/genesis.json" > tmp.json && mv tmp.json "$CHAIN_HOME/config/genesis.json"

echo "Add denom metadata into genesis..."
jq --arg base "$DENOM" --arg disp "${DENOM#u}" '
  .app_state.bank.denom_metadata = [
    {
      description: "Native coin on Void Blockchain",
      denom_units: [
        { denom: $base, exponent: 0, aliases: ["microvoid"] },
        { denom: $disp, exponent: 6, aliases: [] }
      ],
      base: $base,
      display: $disp,
      name: "Void Coin",
      symbol: "VOID",
      uri: "",
      uri_hash: ""
    }
  ]
' "$CHAIN_HOME/config/genesis.json" > tmp.json && mv tmp.json "$CHAIN_HOME/config/genesis.json"


echo "👤 Generate key & add genesis account..."
$BINARY keys add validator --home "$CHAIN_HOME"
echo "Add genesis account..."
$BINARY genesis add-genesis-account validator 100000000000uvoid --home "$CHAIN_HOME"

echo "📝 Generate gentx..."
$BINARY genesis gentx validator 50000000uvoid \
  --chain-id "$CHAIN_ID" --home "$CHAIN_HOME"

echo "📄 Collect gentx & validasi genesis..."
$BINARY genesis collect-gentxs --home "$CHAIN_HOME"
$BINARY genesis validate-genesis --home "$CHAIN_HOME"

echo "✅ Start the chain!"
echo "   Running: "
echo "     $BINARY start --home $CHAIN_HOME"
