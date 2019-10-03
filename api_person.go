package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

///
type VisitShortInfo struct {
	Id int    `json:"id"`
	Dt string `json:"dt"`
}

type PrpShortInfo struct {
	Id        int              `json:"id"`
	Num       string           `json:"num"`
	DtBeg     string           `json:"dtbeg"`
	DtEnd     string           `json:"dtend"`
	VisitList []VisitShortInfo `json:"visits"`
}

type PersonInfo struct {
	Id      int            `json:"id"`
	Fio     string         `json:"fio"`
	Ndoc    string         `json:"ndoc"`
	PrpList []PrpShortInfo `json:"prps"`
}

func JsonApiPersonInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var res DBResult
	idd, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 32)
	if err != nil {
		idd = 0
	}
	res = dbkeeper.GetPersonInfo(int(idd))
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiPersonInfo: error %v", err)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		log.Printf("JsonApiPersonInfo: write error %v", err)
		return
	}
}

func (dk *DBKeeper) GetPersonInfo(id int) DBResult {
	var resultError error = nil
	var resultData PersonInfo
	resultData.PrpList = make([]PrpShortInfo, 0)
	rows, err := dk.db.Query(`select id, fio, ndoc from person where id=$1;`, id)
	if err != nil {
		log.Printf("DBKeeper.GetPersonInfo(1): select query error: %v\n", err)
		return DBResult{err, nil}
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&resultData.Id, &resultData.Fio, &resultData.Ndoc)
		if err != nil {
			log.Printf("DBKeeper.GetPersonInfo(1): get row error: %v\n", err)
			resultError = err
		}
	}
	/////////////
	rows, err = dk.db.Query(`select id, num, dtbeg, dtend from prp where id_own = $1 order by dtbeg desc;`, id)
	if err != nil {
		log.Printf("DBKeeper.GetPersonInfo(2): select query error: %v\n", err)
		return DBResult{err, nil}
	}
	defer rows.Close()
	for rows.Next() {
		var tmp PrpShortInfo
		tmp.VisitList = make([]VisitShortInfo, 0)
		var dt1, dt2 time.Time
		err = rows.Scan(&tmp.Id, &tmp.Num, &dt1, &dt2)
		if err != nil {
			log.Printf("DBKeeper.GetPersonInfo(2): get row error: %v\n", err)
			resultError = err
		}
		tmp.DtBeg = fmt.Sprintf("%02d.%02d.%04d", dt1.Day(), dt1.Month(), dt1.Year())
		tmp.DtEnd = fmt.Sprintf("%02d.%02d.%04d", dt2.Day(), dt2.Month(), dt2.Year())
		resultData.PrpList = append(resultData.PrpList, tmp)
	}
	///////////
	rows, err = dk.db.Query(`select id, id_prp, dt from visit where id_own = $1 order by dt desc;`, id)
	if err != nil {
		log.Printf("DBKeeper.GetPersonInfo(3): select query error: %v\n", err)
		return DBResult{err, nil}
	}
	defer rows.Close()
	for rows.Next() {
		var tmp VisitShortInfo
		var idPrp int
		var dt time.Time
		err = rows.Scan(&tmp.Id, &idPrp, &dt)
		if err != nil {
			log.Printf("DBKeeper.GetPersonInfo(3): get row error: %v\n", err)
			resultError = err
		}
		tmp.Dt = fmt.Sprintf("%02d.%02d.%04d", dt.Day(), dt.Month(), dt.Year())
		for i := range resultData.PrpList {
			if resultData.PrpList[i].Id == idPrp {
				resultData.PrpList[i].VisitList = append(resultData.PrpList[i].VisitList, tmp)
			}
		}
	}
	/////////
	return DBResult{resultError, resultData}
}
