package belajargolang

import "net/http"

func main() {
	println("server berjalan di port:8080")
	http.ListenAndServe(":8080", nil)
}