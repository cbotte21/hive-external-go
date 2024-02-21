package handler

import (
	"context"
	"errors"
	"fmt"
	judicial "github.com/cbotte21/judicial-go/pb"
	"github.com/cbotte21/microservice-common/pkg/datastore"
	"github.com/cbotte21/microservice-common/pkg/jwtParser"
	"github.com/gorilla/websocket"
	"hive-external-go/schema"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func gracefullyDisconnect(userClient *datastore.RedisClient[schema.ActiveUser], conn *websocket.Conn, user schema.ActiveUser) {
	_ = conn.Close()
	//Remove from redis
	if user.Id != "" {
		_ = userClient.Delete(user)
	}
}

func handleKicks(userClient *datastore.RedisClient[schema.ActiveUser], kicked *bool, _id string) {
	sub := userClient.Subscribe("kicks")
	ch := sub.Channel()
	for msg := range ch {
		if msg.Payload == _id {
			*kicked = true
			break
		}
	}
}

func login(userClient *datastore.RedisClient[schema.ActiveUser], user schema.ActiveUser) error {
	return userClient.Create(user)
}

func integrity(judicialClient *judicial.JudicialServiceClient, _id string) error {
	integrity, err := (*judicialClient).Integrity(context.Background(), &judicial.IntegrityRequest{XId: _id})
	if err != nil {
		return err
	}
	if !integrity.Status {
		return errors.New("player is banned")
	}
	return nil
}

func Websocket(w http.ResponseWriter, r *http.Request, userClient *datastore.RedisClient[schema.ActiveUser], judicialClient *judicial.JudicialServiceClient, jwtRedeemer *jwtParser.JwtSecret) {
	// Session variables
	kicked := false
	user := schema.ActiveUser{
		Id:   "",
		Role: 0,
	}

	// Configure socket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer gracefullyDisconnect(userClient, conn, user)

	// Act on message
	for {
		_, p, err := conn.ReadMessage() // Maybe handle messageType?
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		fmt.Printf("Authentication request: <jwt> %s\n", p) // TODO: DEBUG

		// Decode jwt
		res, err := jwtRedeemer.Redeem(string(p))
		if err != nil {
			break
		}
		user.Id = res.Id
		user.Role = res.Role

		// Check player integrity
		err = integrity(judicialClient, user.Id)
		if err != nil {
			break
		}

		// Login
		err = login(userClient, user)
		if err != nil {
			break
		}

		//Handle kicks
		go handleKicks(userClient, &kicked, user.Id)
		for !kicked {
		}
		break
	}
}

//conn.WriteMessage(messageType, p)
