module controller

go 1.23

replace configuration => ./configuration

replace file => ./file

replace service => ./service

require (
	configuration v0.0.0-00010101000000-000000000000
	file v0.0.0-00010101000000-000000000000
	service v0.0.0-00010101000000-000000000000
)
