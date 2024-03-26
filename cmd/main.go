package main

import (
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
	environment.VerifyEnvVariable("port")
	environment.VerifyEnvVariable("judicial_addr")
	environment.VerifyEnvVariable("keycloak_url")
	environment.VerifyEnvVariable("keycloak_realm")

	port := environment.GetEnvVariable("port")
	keycloakUrl := environment.GetEnvVariable("keycloak_url")
	keycloakRealm := environment.GetEnvVariable("keycloak_realm")

	userClient := datastore.RedisClient[schema.ActiveUser]{}
	judicialClient := judicial.NewJudicialServiceClient(getJudicialConn())
	tokenParser := jwtParser.NewJwtParser(keycloakRealm, keycloakUrl)
	userClient.Init()

	api, _ := service.NewApi(port, &judicialClient, &userClient, tokenParser)

	_ = api.Start()
}

func getJudicialConn() *grpc.ClientConn {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(environment.GetEnvVariable("judicial_addr"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf(err.Error())
	}
	return conn
}
