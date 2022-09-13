module github.com/Kong/kong-cloudtrails-integration

go 1.18

require (
	github.com/aws/aws-lambda-go v1.32.0
	github.com/aws/aws-sdk-go v1.44.94
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-resty/resty/v2 v2.7.0
)

replace github.com/aws/aws-sdk-go v1.44.94 => ./aws-sdk-go@v0.1.0-unpublished

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redismock/v8 v8.0.6 // indirect
	github.com/stretchr/objx v0.1.0 // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jarcoal/httpmock v1.2.0
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.2
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
