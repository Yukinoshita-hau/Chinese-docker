module agent

go 1.23

require (
	file v0.0.0-00010101000000-000000000000
	isolated v0.0.0-00010101000000-000000000000
)

require (
	configuration v0.0.0-00010101000000-000000000000
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.9.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/sys v0.28.0 // indirect
)

replace file => ./file

replace configuration => ./configuration

replace isolated => ./isolated
