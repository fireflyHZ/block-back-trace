module profit-allocation

go 1.14

require (
	github.com/astaxie/beego v1.12.3
	github.com/beego/beego/v2 v2.0.1
	github.com/fatih/color v1.9.0
	github.com/filecoin-project/go-address v0.0.5-0.20201103152444-f2023ef3f5bb
	github.com/filecoin-project/go-crypto v0.0.0-20191218222705-effae4ea9f03
	github.com/filecoin-project/go-fil-markets v1.0.10
	github.com/filecoin-project/go-state-types v0.0.0-20201102161440-c8033295a1fc
	github.com/filecoin-project/lotus v1.4.0
	github.com/filecoin-project/specs-actors v0.9.13
	github.com/filecoin-project/specs-actors/v2 v2.3.3
	//github.com/filecoin-project/specs-actors/actors/builtin/reward v0.9.8
	github.com/go-sql-driver/mysql v1.5.0
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-ipld-cbor v0.0.5
	github.com/ipfs/go-log/v2 v2.1.2-0.20200626104915-0016c0b4b3e4
	github.com/prometheus/common v0.10.0
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18 // indirect
	github.com/smartystreets/goconvey v1.6.4
	github.com/urfave/cli v1.22.1
	github.com/urfave/cli/v2 v2.2.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20200826160007-0b9f6c5fb163
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi

//replace github.com/filecoin-project/specs-actors => github.com/filecoin-project/specs-actors v0.9.8
