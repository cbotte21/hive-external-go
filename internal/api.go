package service

import (
	judicial "github.com/cbotte21/judicial-go/pb"
	"github.com/cbotte21/microservice-common/pkg/datastore"
	"github.com/cbotte21/microservice-common/pkg/jwtParser"
	"github.com/gorilla/mux"
	"hive-external-go/internal/handler"
	"hive-external-go/schema"
	"net/http"
)

type Api struct {
	port           string
	router         *mux.Router
	judicialClient *judicial.JudicialServiceClient
	userClient     *datastore.RedisClient[schema.ActiveUser]
	jwtParser      *jwtParser.JwtParser
}

func NewApi(port string, judicialClient *judicial.JudicialServiceClient, userClient *datastore.RedisClient[schema.ActiveUser], jwtParser *jwtParser.JwtParser) (*Api, bool) {
	api := &Api{}
	api.port = port
	api.judicialClient = judicialClient
	api.userClient = userClient
	api.jwtParser = jwtParser
	api.router = mux.NewRouter()
	api.registerHandlers()
	return api, true
}

func (api *Api) Start() error { //maybe change return to bool
	return http.ListenAndServe(":"+api.port, api.router)
}

func (api *Api) registerHandlers() { //Add all API handlers here
	api.router.HandleFunc("/", handler.Status)
	api.router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handler.Websocket(w, r, api.userClient, api.judicialClient, api.jwtParser)
	})
}
