package handlers

import (
	"encoding/json"
	"errors"
	"mychat-backend/database"
	"net/http"
	"gorm.io/gorm"
)

// ユーザー登録用のリクエストを受け取る構造体
type UserRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

func (s *Server) UsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POSTのみ許可されています", http.StatusMethodNotAllowed)
		return
	}

	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "入力データの読み込みに失敗しました", http.StatusBadRequest)
		return
	}

	var user database.User
	err := s.DB.Where("name = ?", req.Name).First(&user).Error

	if err != nil {
		// 💡 判定のポイント！
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// ⭕ 本当に「存在しない」ので、安全に新規登録してOK！
			user = database.User{Name: req.Name, PublicKey: req.PublicKey}
			if err := s.DB.Create(&user).Error; err != nil {
				http.Error(w, "データベースへの保存に失敗しました", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(user)
			return
		}

		// ❌ 「レコードがない」以外の重大なDBエラー（接続切れなど）はここで即座に弾く
		http.Error(w, "データベースエラーが発生しました", http.StatusInternalServerError)
		return
	}

	// 🙅 エラーがない ＝ すでに同じ名前のユーザーが存在する（重複）
	http.Error(w, "このユーザー名はすでに使用されています", http.StatusConflict) // 409 Conflict
}

func (s *Server) SearchUserHandler(w http.ResponseWriter, r *http.Request) {
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
