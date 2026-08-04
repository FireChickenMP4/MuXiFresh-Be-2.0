package main

import (
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/config"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/handler"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/common/nacos"
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/rest"
)

// var configFile = flag.String("f", "etc/task.yaml", "the config file")
func main() {
	flag.Parse()

	var c config.Config
	nacos.MustLoad(nacos.LoadOption{
		Group:  "PROD",
		DataId: "task",
		Target: &c,
	})

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
