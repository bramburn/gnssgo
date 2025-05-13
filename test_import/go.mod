module test_import

go 1.21

replace github.com/bramburn/gnssgo => ../

require github.com/bramburn/gnssgo v0.0.0-00010101000000-000000000000

require (
	github.com/creack/goselect v0.1.2 // indirect
	go.bug.st/serial v1.6.1 // indirect
	golang.org/x/sys v0.15.0 // indirect
)
