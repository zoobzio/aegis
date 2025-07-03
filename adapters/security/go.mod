module aegis/adapters/security

go 1.23.1

replace aegis/catalog => ../../catalog

replace aegis/pipz => ../../pipz

replace aegis/sctx => ../../sctx

replace aegis/capitan => ../../capitan

replace aegis/cereal => ../../cereal

replace aegis/zlog => ../../zlog

require (
	aegis/capitan v0.0.0-00010101000000-000000000000
	aegis/catalog v0.0.0-00010101000000-000000000000
	aegis/sctx v0.0.0-00010101000000-000000000000
)

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	aegis/cereal v0.0.0-00010101000000-000000000000 // indirect
	aegis/pipz v0.0.0-00010101000000-000000000000 // indirect
	aegis/zlog v0.0.0-00010101000000-000000000000 // indirect
)
