package main

import (
	"encoding/json"
	fmt "fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Ack appends to the end of the prev Msg
type ackmsg struct {
	Msg string `json:"msg"`
}

func ack(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got")
	w.Header().Set("Content-Type", "application/json")
	var reqmsg ackmsg
	_ = json.NewDecoder(r.Body).Decode(&reqmsg)
	ackmsg := ackmsg{
		Msg: fmt.Sprintf("%s:ack-sndserv", reqmsg.Msg),
	}
	json.NewEncoder(w).Encode(ackmsg)
}

func main() {
	fmt.Println("hello world")
	r := mux.NewRouter()
	r.HandleFunc("/ack", ack).Methods("POST")
	log.Fatal(http.ListenAndServe("0.0.0.0:5000", r))
}
