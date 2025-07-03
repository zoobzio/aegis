module aegis/moisten

go 1.23.1

require (
	aegis/adapters/security v0.0.0
	aegis/capitan v0.0.0
	aegis/catalog v0.0.0
	aegis/pipz v0.0.0
	aegis/zlog v0.0.0
)

require (
	aegis/cereal v0.0.0-00010101000000-000000000000 // indirect
	aegis/sctx v0.0.0-00010101000000-000000000000 // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace aegis/adapters/security => ../adapters/security

replace aegis/capitan => ../capitan

replace aegis/catalog => ../catalog

replace aegis/cereal => ../cereal

replace aegis/pipz => ../pipz

replace aegis/sctx => ../sctx

replace aegis/zlog => ../zlog
