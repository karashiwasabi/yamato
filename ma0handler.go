// File: ma0handler.go
package main

import (
	"YAMATO/ma0"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

// assuming global DB *sql.DB exists

// 例: productNameHandler を main.go 内に記述する場合

func productNameHandler(w http.ResponseWriter, r *http.Request) {
	yj := r.URL.Query().Get("yj")
	if yj == "" {
		http.Error(w, "yj を指定してください", http.StatusBadRequest)
		return
	}

	// ma0 テーブルから MA018JC018ShouhinMei を取得するクエリ
	const sqlq = `
      SELECT MA018JC018ShouhinMei
      FROM ma0
      WHERE MA009JC009YJCode = ?
      LIMIT 1
    `
	var productName string
	if err := ma0.DB.QueryRow(sqlq, yj).Scan(&productName); err != nil {
		if err == sql.ErrNoRows {
			productName = "" // 存在しなければ空文字
		} else {
			log.Println("productName query error:", err)
			http.Error(w, "DBエラー", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{"productName": productName})
}
