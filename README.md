# Mine-back-trace 
是filecoin出块回测的一个工具，用于检测制定miner在制定高度是否发生丢快，以及出块预测等……
### 创建
go build -o ticket main.go
### 环境变量
```
//指定结果写入的文件位置
export OUTPUT_FILE="./stats.out"
//lotus                                                                               
export LOTUS_HOST="http://ip:port/rpc/v0"
//lotus api sign权限的token，用于计算VRF
export LOTUS_SIGN_TOKEN="eyJhbGciOiJIUzI1NiIsInR..........."
```

### 运行
#### 根据时间
```azure
./ticket stats time --miner f0419945 --start 2021-05-24T00:00:00 --end 2021-05-24T16:21:00
```
#### 根据高度
```azure
./ticket stats epoch --miner=f0419945 --start 780000 --end 784000
```
