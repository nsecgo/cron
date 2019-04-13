package worker

import (
	"context"
	"github.com/nsecgo/crontab/common"
	"go.etcd.io/etcd/clientv3"
	"net"
	"time"
)

// 注册节点到etcd： /cron/workers/IP地址
type Register struct {
	client *clientv3.Client

	localIP string // 本机IP
}

var (
	G_register *Register
)

// 获取本机网卡IP
func getLocalIP() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet // IP地址
		isIpNet bool
	)
	// 获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	// 取第一个非lo的网卡IP
	for _, addr = range addrs {
		// 这个网络地址是IP地址: ipv4, ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String() // 192.168.1.1
				return
			}
		}
	}

	err = common.ERR_NO_LOCAL_IP_FOUND
	return
}

// 注册到/cron/workers/IP, 并自动续租
func (register *Register) keepOnline() {
	var (
		regKey         string
		leaseGrantResp *clientv3.LeaseGrantResponse
		err            error
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)

	for {
		// 注册路径
		regKey = common.JOB_WORKER_DIR + register.localIP

		cancelFunc = nil

		// 创建租约
		if leaseGrantResp, err = register.client.Lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}

		// 注册到etcd
		if _, err = register.client.KV.Put(context.TODO(), regKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			goto RETRY
		}

		cancelCtx, cancelFunc = context.WithCancel(context.Background())

		// 自动续租
		if keepAliveChan, err = register.client.Lease.KeepAlive(cancelCtx, leaseGrantResp.ID); err != nil {
			goto RETRY
		}

		// 处理续租应答
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil { // 续租失败
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}

func InitRegister() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		localIp string
	)

	// 初始化配置
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,                                     // 集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, // 连接超时
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	// 本机IP
	if localIp, err = getLocalIP(); err != nil {
		return
	}

	G_register = &Register{
		client:  client,
		localIP: localIp,
	}

	// 服务注册
	go G_register.keepOnline()

	return
}
