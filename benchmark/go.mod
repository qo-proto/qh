module github.com/qh-project/qh/benchmark

go 1.25.1

require (
	github.com/qh-project/qh v0.0.0
	github.com/quic-go/qpack v0.5.1
	github.com/quic-go/quic-go v0.56.0
	golang.org/x/net v0.47.0
)

require (
	github.com/MatusOllah/slogcolor v1.7.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/tbocek/qotp v0.2.2 // indirect
	golang.org/x/crypto v0.44.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)

// TODO: don't replace and use actual repo once ready
//nolint:gomoddirectives // Local replacement needed for now
replace github.com/qh-project/qh => ../
