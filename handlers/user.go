package handlers

import (
	"encoding/json"
	"net/http"

	"mychat-backend/database"
)

// ユーザー登録用のリクエストを受け取る構造体
type UserRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

func (s *Server)UsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req UserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "入力データの読み込みに失敗しました", http.StatusBadRequest)
			return
		}

		var user database.User
		// Nameで検索して、無ければ新規作成、あれば公開鍵を「上書き」する
		result := s.DB.Where("name = ?", req.Name).First(&user)
		if result.Error != nil {
			// 見つからない場合は、新しいユーザーとして登録
			user = database.User{Name: req.Name, PublicKey: req.PublicKey}
			s.DB.Create(&user)
		} else {
			// 見つかった場合（再ログイン時）は、公開鍵を最新のものに更新する
			user.PublicKey = req.PublicKey
			s.DB.Save(&user)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	} else {
		http.Error(w, "許可されていないメソッドです", http.StatusMethodNotAllowed)
	}
}

func (s *Server)SearchUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// URLのパラメータから名前を取得（例: /users/search?name=太郎）
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "名前が指定されていません", http.StatusBadRequest)
			return
		}

		var user database.User
		result := s.DB.Where("name = ?", name).First(&user)
		if result.Error != nil {
			http.Error(w, "ユーザーが見つかりません", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	} else {
		http.Error(w, "許可されていないメソッドです", http.StatusMethodNotAllowed)
	}
}