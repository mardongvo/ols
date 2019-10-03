package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type PrpRow struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type PrpInfo struct {
	Id         int      `json:"id"`
	Num        string   `json:"num"`
	DtBeg      string   `json:"dtbeg"`
	DtEnd      string   `json:"dtend"`
	PersonId   int      `json:"person_id"`
	PersonFio  string   `json:"person_fio"`
	PersonNdoc string   `json:"person_ndoc"`
	Rows       []PrpRow `json:"rows"`
}

type PrpSaveRequest struct {
	Id   int      `json:"id"` //prp id
	Rows []PrpRow `json:"rows"`
}

func JsonApiPrpInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var res DBResult
	idd, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 32)
	if err != nil {
		idd = 0
	}
	res = dbkeeper.GetPrpInfo(int(idd))
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiPrpInfo: error %v", err)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		log.Printf("JsonApiPrpInfo: write error %v", err)
		return
	}
}

func JsonApiPrpSave(w http.ResponseWriter, r *http.Request) {
	var res DBResult
	var rq PrpSaveRequest
	inp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("JsonApiPrpSave(1): error %v", err)
		res = DBResult{fmt.Errorf("%v", err), nil}
		goto END
	}
	err = json.Unmarshal(inp, &rq)
	if err != nil {
		log.Printf("JsonApiPrpSave(2): error %v", err)
		res = DBResult{fmt.Errorf("%v", err), nil}
		goto END
	}
	err = dbkeeper.SavePrpInfo(rq)
	if err != nil {
		log.Printf("JsonApiPrpSave(3): error %v", err)
		res = DBResult{fmt.Errorf("%v", err), nil}
		goto END
	}
	res = dbkeeper.GetPrpInfo(rq.Id)
END:
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiPrpSave(4): error %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Printf("JsonApiPrpSave: write error %v", err)
		return
	}
}

func (dk *DBKeeper) GetPrpInfo(id int) DBResult {
	var resultError error = nil
	var resultData PrpInfo
	resultData.Rows = make([]PrpRow, 0)
	/////////
	rows, err := dk.db.Query(`select prp.id, prp.num, prp.dtbeg, prp.dtend,
	 person.id, person.fio, person.ndoc from person, prp where prp.id=$1
	 and prp.id_own=person.id;`, id)
	if err != nil {
		log.Printf("DBKeeper.GetPrpInfo(1): select query error: %v\n", err)
		return DBResult{err, nil}
	}
	defer rows.Close()
	for rows.Next() {
		var dt1, dt2 time.Time
		err = rows.Scan(&resultData.Id, &resultData.Num, &dt1, &dt2,
			&resultData.PersonId, &resultData.PersonFio, &resultData.PersonNdoc)
		if err != nil {
			log.Printf("DBKeeper.GetPrpInfo(1): get row error: %v\n", err)
			resultError = err
		}
		resultData.DtBeg = fmt.Sprintf("%02d.%02d.%04d", dt1.Day(), dt1.Month(), dt1.Year())
		resultData.DtEnd = fmt.Sprintf("%02d.%02d.%04d", dt2.Day(), dt2.Month(), dt2.Year())
	}
	///////
	rows, err = dk.db.Query(`select id, name, cnt from prp_template where id_own=$1 order by lower(name);`, id)
	if err != nil {
		log.Printf("DBKeeper.GetPrpInfo(2): select query error: %v\n", err)
		return DBResult{err, nil}
	}
	defer rows.Close()
	for rows.Next() {
		var tmp PrpRow
		err = rows.Scan(&tmp.Id, &tmp.Name, &tmp.Count)
		if err != nil {
			log.Printf("DBKeeper.GetPrpInfo(2): get row error: %v\n", err)
			resultError = err
		}
		resultData.Rows = append(resultData.Rows, tmp)
	}
	///////
	return DBResult{resultError, resultData}
}

func (dk *DBKeeper) SavePrpInfo(rq PrpSaveRequest) error {
	tx, err := dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.SavePrpInfo(1): begin tx error: %v\n", err)
		return err
	}
	for _, v := range rq.Rows {
		if v.Id > 0 {
			_, err := tx.Exec("update prp_template set name=$1, cnt=$2 where id=$3 and id_own=$4;",
				v.Name, v.Count, v.Id, rq.Id)
			if err != nil {
				tx.Rollback()
				log.Printf("DBKeeper.SavePrpInfo(2): update error: %v\n", err)
				return err
			}
		} else {
			if v.Name == "" || v.Count == 0 {
				continue
			}
			_, err := tx.Exec(`insert into prp_template(name,cnt,id_own)
			 select $1::varchar, $2, $3 where not exists (select id from prp_template where
			name=$1 and id_own=$3);`,
				v.Name, v.Count, rq.Id)
			if err != nil {
				tx.Rollback()
				log.Printf("DBKeeper.SavePrpInfo(3): insert error: %v\n", err)
				return err
			}
		}
	}
	return tx.Commit()
}
