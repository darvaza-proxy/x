module darvaza.org/x/tls

go 1.21

require (
	darvaza.org/core v0.15.2
	darvaza.org/slog v0.5.14
)

require (
	darvaza.org/x/container v0.0.0-00010101000000-000000000000
	github.com/zeebo/blake3 v0.2.4
	golang.org/x/crypto v0.29.0
)

require (
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.20.0 // indirect
)

replace darvaza.org/x/container => ../container
