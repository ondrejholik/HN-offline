package main

import (
	"html/template"
	"log"
	"os"
)

const ()

func main() {

	const (
		path = "templates/index.tmpl"
	)

	f, err := os.Create("index.html")
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	t, err := template.ParseFiles(path)
	items := FetchTop()

	err = t.Execute(f, items)
	if err != nil {
		panic(err)
	}

}
