package main

import (
	"MuXiFresh-Be-2.0/common/nacos"
	"flag"
	"fmt"

	"MuXiFresh-Be-2.0/app/form/api/internal/config"
	"MuXiFresh-Be-2.0/app/form/api/internal/handler"
	"MuXiFresh-Be-2.0/app/form/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/form-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	nacos.MustLoad(nacos.LoadOption{
		Group:  "PROD",
		DataId: "form-api",
		Target: &c,
	})

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
