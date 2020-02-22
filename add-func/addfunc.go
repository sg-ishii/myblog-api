package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
)

func handler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "https://app.sg-ishii.page")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	// POSTのみ許可
	if r.Method != http.MethodPost {
		log.Fatalln("method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed) // 405
		return
	}

	// パラメータを取得
	err := r.ParseForm()
	if err != nil {
		log.Fatalln(err)
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	uuid := r.Form.Get("uuid")
	slag := r.Form.Get("slag")

	// パラメータエラー
	if uuid == "" || slag == "" {
		log.Fatalln("invalid param")
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// 環境変数からプロジェクトIDを取得
	projectID := os.Getenv("PROJECT_ID")

	// デフォルトサービスアカウントの証明書を設定
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	defer client.Close()

	// Firestoreにデータを登録
	doc := make(map[string]interface{})
	doc[slag] = true
	_, err = client.Collection("reads").Doc(uuid).Set(ctx, doc, firestore.MergeAll)
	if err != nil {
		log.Fatalln(err)
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}

	// Cookieにuuidを設定
	cookie := &http.Cookie{
		Name:  "uuid",
		Value: uuid,
	}
	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
