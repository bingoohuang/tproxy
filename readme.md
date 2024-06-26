# tproxy

English | [简体中文](readme-cn.md)

![img.png](images/2023-07-06.png)

<a href="https://www.buymeacoffee.com/bingoohuang" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;" ></a>

## Why I wrote this tool

When I develop backend services and write [go-zero](https://github.com/zeromicro/go-zero), I often need to monitor the network traffic. For example:
1. monitoring gRPC connections, when to connect and when to reconnect
2. monitoring MySQL connection pools, how many connections and figure out the lifetime policy
3. monitoring any TCP connections on the fly

## Installation

```shell
$ go install github.com/bingoohuang/tproxy@latest
```

Or use docker images:

```shell
$ docker run --rm -it -p <listen-port>:<listen-port> -p <remote-port>:<remote-port> kevinwan/tproxy:v1 tproxy -l 0.0.0.0 -p <listen-port> -r host.docker.internal:<remote-port>
```

For arm64:

```shell
$ docker run --rm -it -p <listen-port>:<listen-port> -p <remote-port>:<remote-port> kevinwan/tproxy:v1-arm64 tproxy -l 0.0.0.0 -p <listen-port> -r host.docker.internal:<remote-port>
```

On Windows, you can use [scoop](https://scoop.sh/):

```shell
$ scoop install tproxy
```

## Usages

```shell
$ tproxy --help
Usage of tproxy:
  -d duration
    	the delay to relay packets
  -down int
    	Downward speed limit(bytes/second)
  -l string
    	Local address to listen on (default "localhost")
  -p int
    	Local port to listen on, default to pick a random port
  -q	Quiet mode, only prints connection open/close and stats, default false
  -r string
    	Remote address (host:port) to connect
  -s	Enable statistics
  -t string
    	The type of protocol, currently support http2, grpc, redis and mongodb
  -up int
    	Upward speed limit(bytes/second)
```

## Examples

### Monitor gRPC connections

```shell
$ tproxy -p 8088 -r localhost:8081 -t grpc -d 100ms
```

- listen on localhost and port 8088
- redirect the traffic to `localhost:8081`
- protocol type to be gRPC
- delay 100ms for each packets

<img width="579" alt="image" src="https://user-images.githubusercontent.com/1918356/181794530-5b25f75f-0c1a-4477-8021-56946903830a.png">

### Monitor MySQL connections

```shell
$ tproxy -p 3307 -r localhost:3306
```

<img width="600" alt="image" src="https://user-images.githubusercontent.com/1918356/173970130-944e4265-8ba6-4d2e-b091-1f6a5de81070.png">

### Check the connection reliability (Retrans rate and RTT)

```shell
$ tproxy -p 3307 -r remotehost:3306 -s -q
```

<img width="548" alt="image" src="https://user-images.githubusercontent.com/1918356/180252614-7cf4d1f9-9ba8-4aa4-a964-6f37cf991749.png">

### Learn the connection pool behaviors

```shell
$ tproxy -p 3307 -r localhost:3306 -q -s
```

<img width="404" alt="image" src="https://user-images.githubusercontent.com/1918356/236633144-9136e415-5763-4051-8c59-78ac363229ac.png">

## Give a Star! ⭐

If you like or are using this project, please give it a **star**. Thanks!

## scripts

- `docker run --name some-mysql -e MYSQL_ROOT_PASSWORD=root -p 3306:3306 -d mysql`
- `docker search docker-oracle-xe-11g`
- `docker run -h oracle --name oracle -d -p 15210:22 -p 15211:1521 -p 15213:8080 deepdiver/docker-oracle-xe-11g`
- `%connect oracle://system:oracle@127.0.0.1:15211/xe`
- `docker run -h oracle --name oracle -d -p 15210:22 -p 15211:1521 -p 15213:8080 epiclabs/docker-oracle-xe-11g`

- https://hub.docker.com/r/pengbai/docker-oracle-xe-11g-r2
- oracle xe 11g r2 with sql initdb and web console
- `docker run -d -p 8080:8080 -p 1521:1521 pengbai/docker-oracle-xe-11g-r2`

## TODO

- [ ] 服务端 `tproxy -p :29200 -P 127.0.0.01:19200 -t http` 
  客户端 poc olivere-es 时 不能正常打印响应体 `olivere-es -U http://127.0.0.1:29200 -n 2 --trace`,
  客户端 gurl 时正常打印 `gurl 'name=@姓名' 'sex=@random(男,女)' 'addr=@地址' 'idcard=@身份证' :29200/person/_doc/@ksuid -prU`