package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

var dbkeeper DBKeeper

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

func UIHandle(filepath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(filepath, "./template/common_header.tmpl")
		if err != nil {
			log.Printf("Ошибка парсинга шаблона (%s): %v", filepath, err)
			fmt.Printf("Ошибка парсинга шаблона (%s): %v\n", filepath, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Template parsing error"))
			return
		}
		err = tmpl.Execute(w, nil)
		if err != nil {
			log.Printf("Ошибка рендеринга шаблона (%s): %v", filepath, err)
			fmt.Printf("Ошибка рендеринга шаблона (%s): %v", filepath, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Template rendering error"))
			return
		}
	}

}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered: %v", r)
		}
	}()
	//config
	cfgfile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Ошибка при открытии файла конфигурации: %v", err)
	}
	cfg, err := ReadConfig(cfgfile)
	cfgfile.Close()
	if err != nil {
		log.Fatalln(err)
	}

	//log
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	if cfg.LogFile > "" {
		logfile, err := os.Create(cfg.LogFile)
		if err != nil {
			log.Fatalf("Невозможно создать файл: %v", err)
		}
		defer logfile.Close()
		log.SetOutput(logfile)
	}

	dbkeeper = NewDBKeeper(cfg.Database)
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("./static/"))))
	//mux.HandleFunc("/api/", JsonApiHandle)
	mux.HandleFunc("/", UIHandle("./template/root.tmpl"))
	mux.HandleFunc("/person", UIHandle("./template/person.tmpl"))
	mux.HandleFunc("/prp", UIHandle("./template/prp.tmpl"))
	mux.HandleFunc("/visit", UIHandle("./template/visit.tmpl"))
	mux.HandleFunc("/farm", UIHandle("./template/farm.tmpl"))
	mux.HandleFunc("/stat", UIHandle("./template/stat.tmpl"))
	mux.HandleFunc("/znvlp", UIHandle("./template/znvlp.tmpl"))

	mux.HandleFunc("/visit.xlsx", ApiGetXls)

	mux.HandleFunc("/api/person_search", JsonApiFindPerson)
	mux.HandleFunc("/api/person_info", JsonApiPersonInfo)
	mux.HandleFunc("/api/prp_info", JsonApiPrpInfo)
	mux.HandleFunc("/api/prp_save", JsonApiPrpSave)
	mux.HandleFunc("/api/visit_info", JsonApiVisitInfo)
	mux.HandleFunc("/api/visit_save", JsonApiVisitSave)
	mux.HandleFunc("/api/visit_remove", JsonApiVisitRemove)
	mux.HandleFunc("/api/farm_list", JsonApiFarmList)
	mux.HandleFunc("/api/farm_candidates", JsonApiFarmCandidates)
	mux.HandleFunc("/api/farm_candidates_save", JsonApiCandidatesSave)
	mux.HandleFunc("/api/farm_add", JsonApiFarmAddNew)
	mux.HandleFunc("/api/stat", JsonApiStatInfo)
	mux.HandleFunc("/api/znvlp", JsonApiZnvlp)
	mux.HandleFunc("/api/znvlp_price", JsonApiZnvlp)

	s := &http.Server{
		Addr:           cfg.ListenAddress,
		Handler:        mux,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
