package main

import (
	"MuXiFresh-Be-2.0/common/nacos"
	"flag"
	"fmt"

	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/config"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/server"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/pb"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/accountCenter.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	nacos.MustLoad(nacos.LoadOption{
		Group:  "PROD",
		DataId: "accountCenter",
		Target: &c,
	})
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterAccountCenterClientServer(grpcServer, server.NewAccountCenterClientServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
