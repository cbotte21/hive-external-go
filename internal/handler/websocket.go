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
	fmt.Println("Graceful disconnect called")
	fmt.Println(user)
	if user.Id != "" {
		err := userClient.Delete(user)
		fmt.Println(err)
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

func Websocket(w http.ResponseWriter, r *http.Request, userClient *datastore.RedisClient[schema.ActiveUser], judicialClient *judicial.JudicialServiceClient, jwtParser *jwtParser.JwtParser) {
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

		// Decode jwt
		res, err := jwtParser.Redeem(string(p))

		user.Id, _ = res.GetSubject()
		user.Role = 0 // TODO: Set based on claim role

		// Check player integrity
		err = integrity(judicialClient, user.Id)
		if err != nil {
			conn.WriteMessage(0, []byte(err.Error()))
			break
		}

		// Login
		err = login(userClient, user)
		if err != nil {
			conn.WriteMessage(0, []byte(err.Error()))
			break
		}

		//Handle kicks
		go handleKicks(userClient, &kicked, user.Id)
		for !kicked && err == nil {
			_, _, err = conn.NextReader()
		}
		break
	}
}

//conn.WriteMessage(messageType, p)
