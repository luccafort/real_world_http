package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/skratchdot/open-golang/open"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var clinetID = os.Getenv("GITHUB_OAUTH_ID")
var clientSecret = os.Getenv("GITHUB_OAUTH_SECRET")
var redirectURL = "https://localhost:18888"
var state = "your state"

func main() {
	conf := &oauth2.Config{
		ClientID:     clinetID,
		ClientSecret: clientSecret,
		Scopes:       []string{"user:email", "gist"},
		Endpoint:     github.Endpoint,
	}
	// 初期化
	var token *oauth2.Token
	filename := "access_token.json"

	// ローカルに保存済み？
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		// 初回アクセス
		// 認可画面のURLを取得
		url := conf.AuthCodeURL(state, oauth2.AccessTypeOnline)

		// コールバックを受け取るWebサーバをセット
		code := make(chan string)
		var server *http.Server
		server = &http.Server{
			Addr: ":18888",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// クエリパラメータからcodeを取得し、ブラウザを閉じる
				w.Header().Set("Content-Type", "text/html")
				io.WriteString(w, "<html><script>window.open('about:blank', self).close()</script></html>")
				w.(http.Flusher).Flush()
				code <- r.URL.Query().Get("code")
				// サーバも閉じる
				server.Shutdown(context.Background())
			}),
		}
		go server.ListenAndServe()

		// ブラウザで認可画面を開く
		// Githubの認可が完了すれば上記のサーバにリダイレクトされてHandlerが実行される
		open.Start(url)

		// 取得したコーdpwpアクセストークンに交換
		token, err := conf.Exchange(oauth2.NoContext, <-code)
		if err != nil {
			panic(err)
		}

		// アクセストークンをファイルに保存
		file, err := os.Create(filename)
		if err != nil {
			panic(err)
		}

		json.NewEncoder(file).Encode(token)
	} else if err == nil {
		token = &oauth2.Token{}
		json.NewDecoder(file).Decode(token)
	} else {
		panic(err)
	}

	client := oauth2.NewClient(oauth2.NoContext, conf.TokenSource(oauth2.NoContext, token))

	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	emails, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(emails))
}
