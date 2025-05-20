module github.com/bramburn/gnssgo/pkg/ntrip

go 1.21

require (
	github.com/bramburn/gnssgo/pkg/gnssgo v0.0.0
	github.com/stretchr/testify v1.10.0
)

replace github.com/bramburn/gnssgo/pkg/gnssgo => ../gnssgo
