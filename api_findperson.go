package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

//Структуры для API
type PersonShort struct {
	Id   int    `json:"id"`
	Fio  string `json:"fio"`
	Ndoc string `json:"ndoc"`
}

func JsonApiFindPerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var res DBResult
	//get key search
	search := r.URL.Query().Get("search")
	//search
	if len(search) < 1 {
		res = DBResult{fmt.Errorf("Длина строки поиска должна быть не меньше 1 символа"), nil}
	} else {
		res = dbkeeper.FindPerson(search)
	}
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiFindPerson: error %v", err)
		return
	}
	w.Write(data)
}

//API
func (dk *DBKeeper) FindPerson(search string) DBResult {
	if search == "" {
		return DBResult{fmt.Errorf("Строка поиска пуста"), nil}
	}
	var resultError error = nil
	var resultData []PersonShort = make([]PersonShort, 0)
	rows, err := dk.db.Query(`select id, fio, ndoc from person where (position(lower($1) in
	 lower(fio) )>0 or ndoc=$2) and active=1 order by lower(fio) limit 100;`, search, search)
	if err != nil {
		log.Printf("DBKeeper.FindPerson: select query error: %v\n", err)
		return DBResult{err, nil}
	}
	defer rows.Close()
	for rows.Next() {
		var tmp PersonShort
		err = rows.Scan(&tmp.Id, &tmp.Fio, &tmp.Ndoc)
		if err != nil {
			log.Printf("DBKeeper.FindPerson: get row error: %v\n", err)
			resultError = err
			//break
		}
		resultData = append(resultData, tmp)
	}
	return DBResult{resultError, resultData}
}
