module github.com/bramburn/gnssgo/examples/ntrip/server

go 1.23

require (
	github.com/bramburn/gnssgo/pkg/caster v0.0.0
	github.com/bramburn/gnssgo/pkg/gnssgo v0.0.0
	github.com/bramburn/gnssgo/pkg/server v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

replace (
	github.com/bramburn/gnssgo/pkg/caster => ../../../pkg/caster
	github.com/bramburn/gnssgo/pkg/gnssgo => ../../../pkg/gnssgo
	github.com/bramburn/gnssgo/pkg/server => ../../../pkg/server
)
