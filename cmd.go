package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

func JsonApiHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()
	res := "{"
	for k, v := range q {
		res += "\"" + k + "\":\"" + v[0] + "\", "
	}
	res += "\"}"
	w.Write([]byte(res))
}

func RootHandle(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./template/root.tmpl")
	if err != nil {
		log.Printf("Ошибка парсинга шаблона (template/root.tmpl): %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Template parsing error"))
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Ошибка рендеринга шаблона (template/root.tmpl): %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Template rendering error"))
		return
	}
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	logfile, err := os.Create("ols.log")
	if err != nil {
		log.Fatalf("Невозможно создать файл: %v", err)
	}
	defer logfile.Close()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered: %v", r)
		}
	}()
	log.SetOutput(logfile)
	cfgfile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Ошибка при открытии файла конфигурации: %v", err)
	}
	cfg, err := ReadConfig(cfgfile)
	cfgfile.Close()
	if err != nil {
		log.Fatalln(err)
	}

	dbkeeper := NewDBKeeper(cfg.Database)
	fmt.Println(dbkeeper.isAlive())

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("./static/"))))
	mux.HandleFunc("/api/", JsonApiHandle)
	mux.HandleFunc("/", RootHandle)

	s := &http.Server{
		Addr:           cfg.ListenAddress,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
