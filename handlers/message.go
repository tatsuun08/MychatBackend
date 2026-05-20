package handlers

import (
	"net/http"
	"fmt"
	"encoding/json"

	"mychat-backend/database"
)

func (s *Server)MessageHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {

    case http.MethodGet:
        var messages []database.Message
        roomIdStr := r.URL.Query().Get("room_id")
        fmt.Println("GET /messages リクエストを受信:", r.URL.Query())

        if roomIdStr != "" {
            s.DB.Where("messages.room_id = ?", roomIdStr).Find(&messages)
        } else {
            fmt.Println("RoomIDを指定してください")
            http.Error(w, "RoomIDを指定してください", http.StatusBadRequest)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(messages)


    case http.MethodPost:
        fmt.Println("POST /messages リクエストを受信しました！")
        var newMessage database.Message
		
        err := json.NewDecoder(r.Body).Decode(&newMessage)
        if err != nil {
            fmt.Println("デコードエラー:", err)
            http.Error(w, "JSONの読み込みに失敗しました", http.StatusBadRequest)
            return
        }

        result := s.DB.Create(&newMessage)
        if result.Error != nil {
            http.Error(w, "データベースへの保存に失敗しました", http.StatusInternalServerError)
            return
        }


        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated) 
        json.NewEncoder(w).Encode(newMessage)

    default:
        http.Error(w, "許可されていないメソッドです", http.StatusMethodNotAllowed)
    }
}