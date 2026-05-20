package handlers

import (
	"fmt"
	"encoding/json"
	"net/http"
	
	"mychat-backend/database"

)

// /rooms にアクセスが来た時の処理
func (s* Server)RoomsHandler(w http.ResponseWriter, r *http.Request) {
	//POST GETで条件分岐
	switch r.Method {
	//RoomEntityの取得
	case http.MethodGet:
		// userID を取り出す！
	userID, ok := r.Context().Value(UserIDKey).(uint)
	if !ok {
		// 万が一入っていなければ、認証されていないとして弾く
		http.Error(w, "ユーザー識別失敗", http.StatusUnauthorized)
		return
	}

	// ログで確認（Android側から送られてきたIDではなく、JWTから解読した本物のIDです）
	fmt.Printf("【識別成功】ユーザーID: %d からの部屋一覧リクエストです\n", userID)

	var rooms []database.Room

	// 💡 取り出した userID を使って安全にDBから検索！
	s.DB.Joins("JOIN room_users ON room_users.room_id = rooms.id").
		Where("room_users.user_id = ?", userID). // 👈 ここに直接指定できる
		Limit(20).
		Find(&rooms)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)

	//新たなRoomEntityの保存
	case http.MethodPost:
		fmt.Println("POST /rooms リクエストを受信しました！")
		var newRoom database.Room
		
		err := json.NewDecoder(r.Body).Decode(&newRoom)
		if err != nil {
			fmt.Println("デコードエラー:", err)
			http.Error(w, "JSONの読み込みに失敗しました", http.StatusBadRequest)
			return
		}

		result := s.DB.Create(&newRoom)
		fmt.Printf("保存するデータ: %+v\n", newRoom) // ★中身を表示
		if result.Error != nil {
			http.Error(w, "データベースへの保存に失敗しました", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		//サーバーで作成したRoomIDを伝えるために返す 
		json.NewEncoder(w).Encode(newRoom)

	default:
		http.Error(w, "許可されていないメソッドです", http.StatusMethodNotAllowed)
	}
}

func (s *Server)RoomUsersListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fmt.Println("GET /room_users/list")
		roomIdStr := r.URL.Query().Get("room_id")
		if roomIdStr == "" {
			http.Error(w, "RoomIDを指定してください", http.StatusBadRequest)
			return
		}

		var users []database.User
		// JOINを使って、room_usersテーブルと紐づいているusersの情報を抽出
		s.DB.Joins("JOIN room_users ON room_users.user_id = users.id").
			Where("room_users.room_id = ?", roomIdStr).
			Find(&users)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	} else {
		http.Error(w, "許可されていないメソッドです", http.StatusMethodNotAllowed)
	}
}

type RoomUserRequest struct {
	RoomID uint `json:"room_id"`
	UserID uint `json:"user_id"`
}

func (s *Server)RoomUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req RoomUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "入力データの読み込みに失敗しました", http.StatusBadRequest)
			return
		}

		var roomUser database.RoomUser
		result := s.DB.Where(database.RoomUser{RoomID: req.RoomID, UserID: req.UserID}).FirstOrCreate(&roomUser)
		if result.Error != nil {
			fmt.Println("DBエラー:", result.Error)
			http.Error(w, "データベース処理に失敗しました", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(roomUser)
	} else {
		http.Error(w, "許可されていないメソッドです", http.StatusMethodNotAllowed)
	}
}