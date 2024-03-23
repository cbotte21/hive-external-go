package main

import (
	"fmt"
	judicial "github.com/cbotte21/judicial-go/pb"
	"github.com/cbotte21/microservice-common/pkg/datastore"
	"github.com/cbotte21/microservice-common/pkg/environment"
	"github.com/cbotte21/microservice-common/pkg/jwtParser"
	"google.golang.org/grpc"
	service "hive-external-go/internal"
	"hive-external-go/schema"
	"log"
)

func main() {
	port := environment.GetEnvVariable("port")

	userClient := datastore.RedisClient[schema.ActiveUser]{}
	judicialClient := judicial.NewJudicialServiceClient(getJudicialConn())
	jwtRedeemer := jwtParser.JwtSecret(environment.GetEnvVariable("jwt_secret"))
	userClient.Init()

	api, _ := service.NewApi(port, &judicialClient, &userClient, &jwtRedeemer)

	_ = api.Start()
}

func getJudicialConn() *grpc.ClientConn {
	var conn *grpc.ClientConn
	fmt.Println(environment.GetEnvVariable("judicial_addr"))
	conn, err := grpc.Dial(environment.GetEnvVariable("judicial_addr"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf(err.Error())
	}
	return conn
}
