module profit-allocation

go 1.14

require (
	github.com/astaxie/beego v1.12.3
	github.com/beego/beego/v2 v2.0.1
	github.com/blinkbean/dingtalk v0.0.0-20210905093040-7d935c0f7e19
	github.com/fatih/color v1.9.0
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-state-types v0.1.1-0.20210506134452-99b279731c48
	github.com/filecoin-project/lotus v1.10.0
	github.com/filecoin-project/specs-actors v0.9.13
	github.com/filecoin-project/specs-actors/v2 v2.3.5
	github.com/filecoin-project/specs-actors/v5 v5.0.1
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-ipld-cbor v0.0.5
	github.com/ipfs/go-log/v2 v2.1.3
	github.com/lib/pq v1.7.0
	github.com/prometheus/common v0.10.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.7.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20210219115102-f37d292932f2
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi

//replace github.com/filecoin-project/specs-actors => github.com/filecoin-project/specs-actors v0.9.8
