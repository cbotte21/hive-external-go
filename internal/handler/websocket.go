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
	"time"
)

/*
FIRST MESSAGE SEND JWT
REST OF MESSAGES SEND STATUS UPDATE
*/

const POLL_TIME_SECONDS = time.Second * 20

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func gracefullyDisconnect(userClient *datastore.RedisClient[schema.ActiveUser], conn *websocket.Conn, user *schema.ActiveUser) {
	_ = conn.Close()
	//Remove from redis
	if user.Id != "" {
		_ = userClient.Delete(*user)
	}
}

func handleKicks(conn *websocket.Conn, userClient *datastore.RedisClient[schema.ActiveUser], kicked *bool, _id string) {
	sub := userClient.Subscribe("kicks")
	ch := sub.Channel()
	for msg := range ch {
		if msg.Payload == _id {
			conn.WriteMessage(0, []byte("Kicked"))
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

func authenticate(user *schema.ActiveUser, jwtParser *jwtParser.JwtParser, judicialClient *judicial.JudicialServiceClient, conn *websocket.Conn, userClient *datastore.RedisClient[schema.ActiveUser], p []byte, kicked *bool) bool {
	// Decode jwt
	res, err := jwtParser.Redeem(string(p))

	user.Id, _ = res.GetSubject()
	user.Role = 1 // TODO: Set based on claim role

	// Check player integrity
	err = integrity(judicialClient, user.Id)
	if err != nil {
		conn.WriteMessage(0, []byte(err.Error()))
		return false
	}

	// Login
	err = login(userClient, *user)
	if err != nil {
		conn.WriteMessage(0, []byte(err.Error()))
		return false
	}
	conn.WriteMessage(0, []byte("Connection established."))
	go handleKicks(conn, userClient, kicked, user.Id)

	return true
}

func Websocket(w http.ResponseWriter, r *http.Request, userClient *datastore.RedisClient[schema.ActiveUser], judicialClient *judicial.JudicialServiceClient, jwtParser *jwtParser.JwtParser) {
	// Session variables
	kicked := false
	user := schema.ActiveUser{
		Id:       "",
		Role:     0,
		Activity: "/auth/signin",
	}

	// Configure socket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer gracefullyDisconnect(userClient, conn, &user)

	// Act on message
	for {
		_, p, err := conn.ReadMessage() // Maybe handle messageType?
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		if kicked {
			break
		}

		if user.Id == "" { // Authenticate
			if !authenticate(&user, jwtParser, judicialClient, conn, userClient, p, &kicked) {
				break
			}
		} else { // Update status
			user.Activity = string(p)
			_ = userClient.Update(user, user)
		}
	}
}

//conn.WriteMessage(messageType, p)
