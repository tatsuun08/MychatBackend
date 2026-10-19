# MychatBackend

E2EE（エンドツーエンド暗号化）チャットアプリのバックエンドサーバーです。
フレームワークを使わず、Go言語の標準ライブラリを用いて低レイヤからAPIを構築しています。

## 技術スタック
- 言語: Go
- 通信: net/http (REST API)
- データベース: PostgreSQL
- バージョン管理: Git

## 主な機能
- ユーザー登録・認証API
- E2EE用の公開鍵登録・鍵交換API
- 暗号化メッセージの中継・送受信API

## クライアントリポジトリ
Androidクライアントのリポジトリはこちら:
https://github.com/tatsuun08/MyChat
