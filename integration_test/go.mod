module github.com/pbdeuchler/assistant-server/integration_test

go 1.24.3

replace github.com/pbdeuchler/assistant-server => ../

require (
	github.com/go-chi/chi/v5 v5.2.2
	github.com/jackc/pgx/v5 v5.7.5
	github.com/pbdeuchler/assistant-server v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.9.0
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/caarlos0/env/v11 v11.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-chi/httplog/v3 v3.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mark3labs/mcp-go v0.37.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
