nextnet
===

[![GoDoc](https://godoc.org/github.com/hdm/nextnet?status.svg)](https://godoc.org/github.com/hdm/nextnet)

nextnet is a pivot point discovery tool written in Go.

## Download

Binary packages are available [here](https://github.com/hdm/nextnet/releases/latest).

## Install

Most folks should download a compiled binary from the releases page. If you have Go installed, you can build your own binary by running:
```
$ go get github.com/hdm/nextnet
```

## Usage

```
Usage: ./nextnet [options] <cidr or ip address> ... <cidr or ip address>

Probes a list of networks for potential pivot points.

Options:

  -h, --help            Show Usage and exit.
  
  -rate int
    	Set the maximum packets per second rate (default 1000)
      
  -version
    	Show the application version
```

### Local Network Discovery
Identify a multi-homed Windows desktop on the local network.

```
$ nextnet 192.168.0.0/24

{"host":"192.168.0.112","port":"137","proto":"udp","probe":"netbios","name":"DESKTOP-H14GTIO","nets":["192.168.10.12","192.168.20.12"],"info":{"domain":"WORKGROUP","hwaddr":"14:dd:a9:e4:10:a0"}}

```

### Fast External Network Scans
Quickly identify multi-homed hosts running Netbios on the internet

```
$ nextnet -rate 10000 114.80.0.0/16 | grep nets

{"host":"114.80.62.194","port":"137","proto":"udp","probe":"netbios","name":"WIN-6F47E00F5JS","nets":["192.168.80.2","192.168.90.8"],"info":{"domain":"WORKGROUP","hwaddr":"b8:2a:72:d6:e6:b7"}}
{"host":"114.80.60.40","port":"137","proto":"udp","probe":"netbios","name":"ESX140-2008G","nets":["192.168.11.40","114.80.60.40"],"info":{"domain":"WORKGROUP","hwaddr":"00:0c:29:03:df:5a"}}
{"host":"114.80.86.222","port":"137","proto":"udp","probe":"netbios","name":"SHOFFICE-ISA","nets":["114.80.86.222"],"info":{"domain":"\u0001\u0002__MSBROWSE__\u0002","hwaddr":"00:0c:29:5f:ba:d1"}}
{"host":"114.80.153.49","port":"137","proto":"udp","probe":"netbios","name":"WIN-0GRAGSOGFGS","nets":["114.80.153.49","172.16.0.157"],"info":{"domain":"WORKGROUP","hwaddr":"90:b1:1c:40:72:31"}}
{"host":"114.80.156.143","port":"137","proto":"udp","probe":"netbios","name":"WIN-E1GEEJQBT45","nets":["114.80.156.143","192.168.60.1"],"info":{"domain":"WORKGROUP","hwaddr":"b0:83:fe:e9:3b:d0"}}
{"host":"114.80.157.110","port":"137","proto":"udp","probe":"netbios","name":"SHWGQ-DBWEB","nets":["112.65.248.9"],"info":{"domain":"WORKGROUP","hwaddr":"00:26:55:1e:8c:04"}}
{"host":"114.80.157.108","port":"137","proto":"udp","probe":"netbios","name":"DATA-FCMB","nets":["112.65.248.8"],"info":{"domain":"WORKGROUP","hwaddr":"00:1f:29:64:e5:f4"}}
{"host":"114.80.157.170","port":"137","proto":"udp","probe":"netbios","name":"DZH-TRS6","nets":["10.10.2.13"],"info":{"domain":"WORKGROUP","hwaddr":"ac:16:2d:7a:ff:f0"}}
{"host":"114.80.157.44","port":"137","proto":"udp","probe":"netbios","name":"WINAD2","nets":["169.254.35.57","114.80.157.44"],"info":{"domain":"WINAD02","hwaddr":"00:50:56:9d:4f:e1"}}
{"host":"114.80.156.99","port":"137","proto":"udp","probe":"netbios","name":"WUHAN","nets":["10.1.1.2","169.254.95.120"],"info":{"domain":"WORKGROUP","hwaddr":"34:40:b5:9e:cf:28"}}
{"host":"114.80.166.28","port":"137","proto":"udp","probe":"netbios","name":"KEDE-MOBILE-131","nets":["114.80.166.28"],"info":{"domain":"WORKGROUP","hwaddr":"00:50:56:94:54:5c"}}
{"host":"114.80.167.219","port":"137","proto":"udp","probe":"netbios","name":"WIN-DB2BC7UU0CM","nets":["192.168.100.191"],"info":{"domain":"WORKGROUP","hwaddr":"00:50:56:94:49:7c"}}
{"host":"114.80.157.135","port":"137","proto":"udp","probe":"netbios","name":"WIN-L3DFEEB","nets":["169.254.54.216"],"info":{"domain":"WORKGROUP","hwaddr":"90:b1:1c:09:61:cc"}}
{"host":"114.80.207.27","port":"137","proto":"udp","probe":"netbios","name":"R420","nets":["192.168.126.1","192.168.72.1","169.254.73.205"],"info":{"domain":"WORKGROUP","hwaddr":"44:a8:42:3f:8a:23"}}
{"host":"114.80.215.81","port":"137","proto":"udp","probe":"netbios","name":"WINDOWS-7III9JS","nets":["114.80.215.81","192.168.108.1","192.168.233.1"],"info":{"domain":"WORKGROUP","hwaddr":"6c:ae:8b:38:51:f3"}}
{"host":"114.80.215.90","port":"137","proto":"udp","probe":"netbios","name":"HYDROGEN","nets":["114.80.215.90","2.0.1.1","1.1.1.1","192.168.118.1"],"info":{"domain":"WORKGROUP","hwaddr":"e4:1f:13:95:ee:c2"}}
{"host":"114.80.222.219","port":"137","proto":"udp","probe":"netbios","name":"WIN-VINFEGJ7HP8","nets":["114.80.222.219"],"info":{"domain":"WORKGROUP","hwaddr":"14:18:77:41:17:25"}}
{"host":"114.80.245.193","port":"137","proto":"udp","probe":"netbios","name":"WINDOWS-OT2WS9T","nets":["114.80.245.193"],"info":{"domain":"WORKGROUP","hwaddr":"a0:d3:c1:f2:3e:26"}}
```
