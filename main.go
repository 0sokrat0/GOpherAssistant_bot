package main

import "net/http"

const TOKEN = "" //Bot token

func main() {
	resp, err := http.Get("https://api.telegram.org/bot%s/%s", TOKEN, MOTHOD)
}
