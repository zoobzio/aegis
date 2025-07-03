module aegis/capitan

go 1.23.1

require (
	aegis/cereal v0.0.0-00010101000000-000000000000
	aegis/pipz v0.0.0-00010101000000-000000000000
	aegis/sctx v0.0.0-00010101000000-000000000000
	aegis/zlog v0.0.0-00010101000000-000000000000
)

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	aegis/catalog v0.0.0-00010101000000-000000000000 // indirect
)

replace aegis/catalog => ../catalog

replace aegis/cereal => ../cereal

replace aegis/pipz => ../pipz

replace aegis/zlog => ../zlog

replace aegis/sctx => ../sctx
