package main

import (
	"github.com/nikovacevic/wiki"
	"net/http"
)

func main() {
	http.HandleFunc("/view/", wiki.MakeHandler(wiki.ViewHandler))
	http.HandleFunc("/edit/", wiki.MakeHandler(wiki.EditHandler))
	http.HandleFunc("/save/", wiki.MakeHandler(wiki.SaveHandler))
	http.ListenAndServe(":8080", nil)
}
