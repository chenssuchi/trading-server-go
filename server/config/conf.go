package config

var Config *Conf

type Conf struct {
	Pg         PgConfig
	Blockchain BlockchainConfig
}

type BlockchainConfig struct {
	ChainId          int    `toml:"chain_id"`
	RPC              string `toml:"rpc_url"`
	IntoTokenAddress string `toml:"into_token_address"`
	TradeAddress     string `toml:"trade_address"`
	PrivateKey       string `toml:"private_key"`
}

type PgConfig struct {
	ConnStr string `toml:"conn_str"`
}
