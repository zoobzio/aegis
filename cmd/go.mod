module aegis/cmd

go 1.23.1

require (
	aegis/capitan v0.0.0
	aegis/catalog v0.0.0
	aegis/cereal v0.0.0
	aegis/moisten v0.0.0
	aegis/pipz v0.0.0
	aegis/sctx v0.0.0
	aegis/zlog v0.0.0
	github.com/alecthomas/chroma v0.10.0
	github.com/fatih/color v1.15.0
	github.com/spf13/cobra v1.8.0
)

require (
	aegis/adapters/security v0.0.0 // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.8.2 // indirect
	golang.org/x/sys v0.6.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace aegis/adapters/security => ../adapters/security

replace aegis/capitan => ../capitan

replace aegis/catalog => ../catalog

replace aegis/cereal => ../cereal

replace aegis/moisten => ../moisten

replace aegis/pipz => ../pipz

replace aegis/sctx => ../sctx

replace aegis/zlog => ../zlog
