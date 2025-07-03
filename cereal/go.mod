module aegis/cereal

go 1.23.1

require (
	github.com/go-playground/validator/v10 v10.17.0
	github.com/pelletier/go-toml/v2 v2.2.4
	gopkg.in/yaml.v3 v3.0.1
	aegis/catalog v0.0.0-00010101000000-000000000000
	aegis/pipz v0.0.0-00010101000000-000000000000
	aegis/sctx v0.0.0-00010101000000-000000000000
	aegis/zlog v0.0.0-00010101000000-000000000000
)

replace aegis/catalog => ../catalog

replace aegis/pipz => ../pipz

replace aegis/sctx => ../sctx

replace aegis/zlog => ../zlog

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
)
