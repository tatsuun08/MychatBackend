package handlers

import (
	"io"
	"net/http"
	"fmt"

	"mychat-backend/database"
)

func (s *Server)PublicKeyHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        // 1. Queryパラメータから user_id を取得
        userIdStr := r.URL.Query().Get("user_id")
        if userIdStr == "" {
            http.Error(w, "UserIDを指定してください", http.StatusBadRequest)
            return
        }

        // 2. Request Body から公開鍵（String）を読み込む
        body, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "ボディの読み込みに失敗しました", http.StatusInternalServerError)
            return
        }
        publicKey := string(body)

        // 3. データベースの更新
        // ModelにUserを指定し、IDが一致するレコードのPublicKeyカラムのみを更新
        err = s.DB.Model(&database.User{}).Where("id = ?", userIdStr).Update("public_key", publicKey).Error
        if err != nil {
            http.Error(w, "データベースの更新に失敗しました", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, "公開鍵を更新しました")
    } else {
        http.Error(w, "許可されていないメソッドです", http.StatusMethodNotAllowed)
    }
}
