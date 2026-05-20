package main

import (
	"fmt"
	"net/http"

	"mychat-backend/database"
	"mychat-backend/handlers"
)


func main() {
 	gormDB := database.InitDB()
	server := &handlers.Server{
		DB: gormDB,
		JWTSecret: []byte("my_super_secret_chat_key_12345"),
		
	}

	http.HandleFunc("/login", server.LoginHandler)
	http.HandleFunc("/users", server.UsersHandler)

	http.HandleFunc("/rooms", server.AuthMiddleware(server.RoomsHandler))
	http.HandleFunc("/room_user", server.AuthMiddleware(server.RoomUserHandler))
	http.HandleFunc("/messages", server.AuthMiddleware(server.MessageHandler))
	http.HandleFunc("/room_users/list", server.AuthMiddleware(server.RoomUsersListHandler))
	http.HandleFunc("/users/search", server.AuthMiddleware(server.SearchUserHandler))
	http.HandleFunc("/users/public_key", server.AuthMiddleware(server.PublicKeyHandler))

	fmt.Println("サーバー起動: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}