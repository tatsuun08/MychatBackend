package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"mychat-backend/database"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POSTのみ許可されています", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "入力データの読み込みに失敗しました", http.StatusBadRequest)
		return
	}

	var user database.User
	err := s.DB.Where("name = ?", req.Name).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 🙅 ログインなので、存在しない場合は「401 Unauthorized」で弾く
			http.Error(w, "ユーザーが見つかりません", http.StatusUnauthorized)
			return
		}
		// その他の重大なDBエラー
		http.Error(w, "データベースエラーが発生しました", http.StatusInternalServerError)
		return
	}

	// 🛡️ 【前回修正分】ログイン（再インストール時など）のたびに、公開鍵を最新のものに上書き！
	user.PublicKey = req.PublicKey
	if err := s.DB.Save(&user).Error; err != nil {
		http.Error(w, "ユーザー情報の更新に失敗しました", http.StatusInternalServerError)
		return
	}

	// 🎫 JWTの作成
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.JWTSecret)
	if err != nil {
		http.Error(w, "トークンの生成に失敗しました", http.StatusInternalServerError)
		return
	}

	// Androidが欲しがっている情報をまとめて返す
	response := LoginResponse{
		Token:    tokenString,
		UserID:   user.ID,
		UserName: user.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 💡 コンテキスト用の鍵（キー）の型を独自に定義する（衝突防止）
type contextKey string

const UserIDKey contextKey = "userID"

func (s *Server) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "トークンがありません", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("予期しない署名アルゴリズム")
			}
			return s.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "無効なトークンです", http.StatusUnauthorized)
			return
		}

		// 💡 【ここからが識別の重要ロジック！】
		// トークンのクレイム（中身）から user_id を取り出す
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userIDFloat, ok := claims["user_id"].(float64) // JWT内の数値はfloat64で解析されます
			if !ok {
				http.Error(w, "不正なユーザーIDです", http.StatusUnauthorized)
				return
			}
			userID := uint(userIDFloat)

			// 💡 リクエストの「コンテキスト（ポケット）」にユーザーIDを入れる
			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			// 💡 ポケットにIDを入れた「新しいリクエスト」を後ろのハンドラーに引き渡す！
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		http.Error(w, "無効なトークンです", http.StatusUnauthorized)
	}
}
