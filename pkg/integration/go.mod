module github.com/bramburn/gnssgo/pkg/integration

go 1.23

require (
	github.com/bramburn/gnssgo/pkg/caster v0.0.0
	github.com/bramburn/gnssgo/pkg/gnssgo v0.0.0
	github.com/bramburn/gnssgo/pkg/server v0.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
)

replace (
	github.com/bramburn/gnssgo/pkg/caster => ../caster
	github.com/bramburn/gnssgo/pkg/gnssgo => ../gnssgo
	github.com/bramburn/gnssgo/pkg/server => ../server
)
