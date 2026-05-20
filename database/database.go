package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
func InitDB() *gorm.DB {
	var db *gorm.DB
	
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

	return db
}