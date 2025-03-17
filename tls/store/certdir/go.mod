module darvaza.org/x/tls/store/certdir

go 1.22

require (
	darvaza.org/core v0.16.1 // indirect
	darvaza.org/slog v0.6.1
)

require (
	darvaza.org/cache/x/simplelru v0.2.1
	darvaza.org/x/tls v0.0.0-00010101000000-000000000000
)

require (
	darvaza.org/x/container v0.2.1 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)

replace darvaza.org/x/tls => ../..
