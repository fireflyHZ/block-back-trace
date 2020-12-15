module profit-allocation

go 1.14

require github.com/astaxie/beego v1.12.1

require (
	github.com/fatih/color v1.9.0
	github.com/filecoin-project/go-address v0.0.4
	github.com/filecoin-project/go-fil-markets v0.9.1
	github.com/filecoin-project/go-state-types v0.0.0-20201003010437-c33112184a2b
	github.com/filecoin-project/lotus v0.10.2
	github.com/filecoin-project/specs-actors v0.9.12
	github.com/filecoin-project/specs-actors/v2 v2.1.0
	//github.com/filecoin-project/specs-actors/actors/builtin/reward v0.9.8
	github.com/go-sql-driver/mysql v1.5.0
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-ipld-cbor v0.0.5-0.20200428170625-a0bd04d3cbdf
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18 // indirect
	github.com/smartystreets/goconvey v1.6.4
	github.com/urfave/cli v1.22.1
	github.com/urfave/cli/v2 v2.2.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20200826160007-0b9f6c5fb163
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/filecoin-project/filecoin-ffi => ./lotus/filecoin-ffi

//replace github.com/filecoin-project/specs-actors => github.com/filecoin-project/specs-actors v0.9.8
