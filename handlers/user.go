package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"mychat-backend/database"
	"net/http"

	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

// ユーザー登録用のリクエストを受け取る構造体
type UserRequest struct {
	Name      string `json:"name"`
	Password  string `json:"password"`
	PublicKey string `json:"public_key"`
	KeyBackup string `json:"key_backup"`
}

// ランダムなソルト（暗号の塩）を生成する関数
func GenerateSalt(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// ユーザー登録（新規アカウント作成）ハンドラー
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

	// 1. このユーザー専用のランダムソルト（16バイト）を生成し、文字列(Base64)に変換
	saltBytes, _ := GenerateSalt(16)
	saltStr := base64.StdEncoding.EncodeToString(saltBytes)

	// 2. ログイン側(auth.go)の []byte(user.Salt) という計算式と完全に合わせるため、
	// 文字列化したソルトを []byte にキャストして Argon2 に投入します
	hashedPassword := argon2.IDKey([]byte(req.Password), []byte(saltStr), 3, 64*1024, 4, 32)
	hashedPasswordStr := base64.StdEncoding.EncodeToString(hashedPassword)

	// 重複するユーザー名がないかDBを検索
	var user database.User
	err := s.DB.Where("name = ?", req.Name).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// ⭕ 同じ名前のユーザーが「存在しない」ので、安全に新規登録を実行
			user = database.User{
				Name:      req.Name,
				Password:  hashedPasswordStr,
				Salt:      saltStr, // 💡 【大修正】ここに生成したソルトをしっかり保存！
				PublicKey: req.PublicKey,
				KeyBackup: req.KeyBackup,
			}
			
			if err := s.DB.Create(&user).Error; err != nil {
				http.Error(w, "データベースへの保存に失敗しました", http.StatusInternalServerError)
				return
			}

			// 成功レスポンスを返却
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(user)
			return
		}

		// 重大なDBエラーは即座に弾く
		http.Error(w, "データベースエラーが発生しました", http.StatusInternalServerError)
		return
	}

	// エラーがない ＝ すでに同じ名前のユーザーが存在する（重複）
	http.Error(w, "このユーザー名はすでに使用されています", http.StatusConflict)
}

// ユーザーの存在を名前で検索するハンドラー
func (s *Server) SearchUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "許可されていないメソッドです", http.StatusMethodNotAllowed)
		return
	}

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
}