package push

import (
	"fmt"
	"net/http"
)

func ServerHome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("req url: ", r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	fmt.Println("server home")
	http.ServeFile(w, r, "./push/home.html")
}
