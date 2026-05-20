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
	}


	http.HandleFunc("/rooms", server.RoomsHandler)
	http.HandleFunc("/users", server.UsersHandler)
	http.HandleFunc("/room_user", server.RoomUserHandler)
	http.HandleFunc("/messages", server.MessageHandler)
	http.HandleFunc("/room_users/list", server.RoomUsersListHandler)
	http.HandleFunc("/users/search", server.SearchUserHandler)
	http.HandleFunc("/users/public_key", server.PublicKeyHandler)

	fmt.Println("サーバー起動: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}