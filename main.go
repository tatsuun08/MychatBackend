package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


var db *gorm.DB

//Remote DB Table 

//RoomEntity
type Room struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
}

//UserEntity
type User struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

//RoomUserEntity N-N table(Room - User)
type RoomUser struct {
	RoomID uint `gorm:"primaryKey" json:"room_id"`
	UserID uint `gorm:"primaryKey" json:"user_id"`

	Room Room `gorm:"foreignKey:RoomID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}


//MessageEntity
type Message struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Text     string `json:"text"`

	RoomID   uint   `json:"room_id"`
	SenderID uint   `json:"sender_id"` 


	Room   Room `gorm:"foreignKey:RoomID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Sender User `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

// データベースの初期設定
func initDB() {
	{//.envファイルを環境変数として設定
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}
	//環境変数の読み込み
	var user string = os.Getenv("DB_USER")
	var password string = os.Getenv("DB_PASSWORD")
	var db_name string = os.Getenv("DB_NAME")
	var port string = os.Getenv("PORT")

	//DB接続用の情報
	dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%s sslmode=disable", user, password, db_name, port)
	
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("データベース接続に失敗しました: " + err.Error())
	}

	//構造体からテーブルを自動作成
	db.AutoMigrate(&Room{}, &User{}, &RoomUser{}, &Message{})
}

// /rooms にアクセスが来た時の処理
func roomsHandler(w http.ResponseWriter, r *http.Request) {
	//POST GETで条件分岐
	switch r.Method {
	//RoomEntityの取得
	case http.MethodGet:
		//URLから特定の情報を取得http://example.com&user_id=5
		userIDStr := r.URL.Query().Get("user_id")

		var rooms []Room

		if userIDStr != "" {
			db.Joins("JOIN room_users ON room_users.room_id = rooms.id").//Join: room_usersとroomでIDが一致するものを結合，ルーム情報とユーザー情報をまとめる
				Where("room_users.user_id = ?", userIDStr).//指定した文字列とroom_users.user_idが一致するものだけを抽出
				Limit(20).
				Find(&rooms)
		} else {
			fmt.Println("UserIDを指定してください")
			return
		}
		//ヘッダーにJSONであると教える
		w.Header().Set("Content-Type", "application/json")
		//JSONでroomsの値を返す
		json.NewEncoder(w).Encode(rooms)

	//新たなRoomEntityの保存
	case http.MethodPost:
		fmt.Println("POST /rooms リクエストを受信しました！")
		var newRoom Room
		
		err := json.NewDecoder(r.Body).Decode(&newRoom)
		if err != nil {
			fmt.Println("デコードエラー:", err)
			http.Error(w, "JSONの読み込みに失敗しました", http.StatusBadRequest)
			return
		}

		result := db.Create(&newRoom)
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

// ユーザー登録用のリクエストを受け取る構造体
type UserRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req UserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "入力データの読み込みに失敗しました", http.StatusBadRequest)
			return
		}

		var user User
		// Nameで検索して、無ければ新規作成、あれば公開鍵を「上書き」する
		result := db.Where("name = ?", req.Name).First(&user)
		if result.Error != nil {
			// 見つからない場合は、新しいユーザーとして登録
			user = User{Name: req.Name, PublicKey: req.PublicKey}
			db.Create(&user)
		} else {
			// 見つかった場合（再ログイン時）は、公開鍵を最新のものに更新する
			user.PublicKey = req.PublicKey
			db.Save(&user)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	} else {
		http.Error(w, "許可されていないメソッドです", http.StatusMethodNotAllowed)
	}
}

func publicKeyHandler(w http.ResponseWriter, r *http.Request) {
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
        err = db.Model(&User{}).Where("id = ?", userIdStr).Update("public_key", publicKey).Error
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

func roomUsersListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fmt.Println("GET /room_users/list")
		roomIdStr := r.URL.Query().Get("room_id")
		if roomIdStr == "" {
			http.Error(w, "RoomIDを指定してください", http.StatusBadRequest)
			return
		}

		var users []User
		// JOINを使って、room_usersテーブルと紐づいているusersの情報を抽出
		db.Joins("JOIN room_users ON room_users.user_id = users.id").
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

func roomUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req RoomUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "入力データの読み込みに失敗しました", http.StatusBadRequest)
			return
		}

		var roomUser RoomUser
		result := db.Where(RoomUser{RoomID: req.RoomID, UserID: req.UserID}).FirstOrCreate(&roomUser)
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

func messageHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {

    case http.MethodGet:
        var messages []Message
        roomIdStr := r.URL.Query().Get("room_id")
        fmt.Println("GET /messages リクエストを受信:", r.URL.Query())

        if roomIdStr != "" {
            db.Where("messages.room_id = ?", roomIdStr).Find(&messages)
        } else {
            fmt.Println("RoomIDを指定してください")
            http.Error(w, "RoomIDを指定してください", http.StatusBadRequest)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(messages)


    case http.MethodPost:
        fmt.Println("POST /messages リクエストを受信しました！")
        var newMessage Message
        
        err := json.NewDecoder(r.Body).Decode(&newMessage)
        if err != nil {
            fmt.Println("デコードエラー:", err)
            http.Error(w, "JSONの読み込みに失敗しました", http.StatusBadRequest)
            return
        }

        result := db.Create(&newMessage)
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

func searchUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// URLのパラメータから名前を取得（例: /users/search?name=太郎）
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "名前が指定されていません", http.StatusBadRequest)
			return
		}

		var user User
		result := db.Where("name = ?", name).First(&user)
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


func main() {
	initDB()

	http.HandleFunc("/rooms", roomsHandler)
	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/room_user", roomUserHandler)
	http.HandleFunc("/messages", messageHandler)
	http.HandleFunc("/room_users/list", roomUsersListHandler)
	http.HandleFunc("/users/search", searchUserHandler)
	http.HandleFunc("/users/public_key", publicKeyHandler)

	fmt.Println("サーバー起動: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}