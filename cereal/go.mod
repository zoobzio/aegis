module aegis/cereal

go 1.23.1

require (
	aegis/catalog v0.0.0-00010101000000-000000000000
	aegis/pipz v0.0.0-00010101000000-000000000000
	aegis/sctx v0.0.0-00010101000000-000000000000
	github.com/BurntSushi/toml v1.5.0
	gopkg.in/yaml.v3 v3.0.1
)

replace aegis/catalog => ../catalog

replace aegis/pipz => ../pipz

replace aegis/sctx => ../sctx
