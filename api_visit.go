package main

import (
	"encoding/json"
	"fmt"

	//"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type VisitRow struct {
	Id             int     `json:"id"` //prp_template id
	Name           string  `json:"name"`
	Count          int     `json:"count"`
	Price          float32 `json:"price"`
	PriceZnvlp     float32 `json:"price_znvlp"`
	Reason         string  `json:"reason"`
	PayDt          string  `json:"paydt"`
	PrpCount       int     `json:"prp_count"`
	PrevCount      int     `json:"prev_count"`       //
	PrevCountSaved int     `json:"prev_count_saved"` //
	RecepCount     int     `json:"recep_count"`
}

type VisitInfo struct {
	Id         int        `json:"id"`
	Dt         string     `json:"dt"`
	PersonId   int        `json:"person_id"`
	PersonFio  string     `json:"person_fio"`
	PersonNdoc string     `json:"person_ndoc"`
	PrpId      int        `json:"prp_id"`
	PrpNum     string     `json:"prp_num"`
	PrpDtBeg   string     `json:"prp_dtbeg"`
	PrpDtEnd   string     `json:"prp_dtend"`
	Rows       []VisitRow `json:"rows"`
}

type VisitSaveRequest struct {
	Id   int        `json:"id"`
	Rows []VisitRow `json:"rows"`
}

func JsonApiVisitInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var res DBResult
	idd, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 32)
	if err != nil {
		idd = 0
	}
	res = dbkeeper.GetVisitInfo(int(idd))
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiVisitInfo: error %v", err)
		return
	}
	w.Write(data)
}

func (dk *DBKeeper) GetVisitInfo(id int) DBResult {
	var resultError error = nil
	var resultData VisitInfo
	resultData.Rows = make([]VisitRow, 0)
	/////////
	rows, err := dk.db.Query(`select visit.id, visit.dt,
	 prp.id, prp.num, prp.dtbeg, prp.dtend,
	 person.id, person.fio, person.ndoc
	 from person, prp, visit where visit.id=$1
	 and visit.id_prp=prp.id and visit.id_own=person.id;`, id)
	if err != nil {
		log.Printf("DBKeeper.GetVisitInfo(1): select query error: %v\n", err)
		return DBResult{err, nil}
	}
	defer rows.Close()
	for rows.Next() {
		var dt, dt1, dt2 time.Time
		err = rows.Scan(&resultData.Id, &dt, &resultData.PrpId, &resultData.PrpNum, &dt1, &dt2,
			&resultData.PersonId, &resultData.PersonFio, &resultData.PersonNdoc)
		if err != nil {
			log.Printf("DBKeeper.GetVisitInfo(1): get row error: %v\n", err)
			resultError = err
		}
		resultData.Dt = fmt.Sprintf("%02d.%02d.%04d", dt.Day(), dt.Month(), dt.Year())
		resultData.PrpDtBeg = fmt.Sprintf("%02d.%02d.%04d", dt1.Day(), dt1.Month(), dt1.Year())
		resultData.PrpDtEnd = fmt.Sprintf("%02d.%02d.%04d", dt2.Day(), dt2.Month(), dt2.Year())
	}
	///////
	rows, err = dk.db.Query(`with
		curinfo as (select * from visit_info where id_own = $1),
		prev_visits as (select a.id id from visit a, visit b where b.id=$1 and b.id_prp=a.id_prp and a.dt<b.dt),
		prev_counts as (select id_prpt, sum(cnt) cnt from visit_info where id_own in (select id from prev_visits) group by id_prpt)
	select pt.id, pt.name, curinfo.cnt, curinfo.price, curinfo.price_znvlp,
    	curinfo.reason, curinfo.paydt, pt.cnt prp_count, curinfo.prevcnt,
    	coalesce(prev_counts.cnt,0) prevcnt
		from prp_template pt 
        inner join curinfo on (pt.id = curinfo.id_prpt) 
        left outer join prev_counts on (pt.id = prev_counts.id_prpt)
        order by lower(pt.name);`, id)
	if err != nil {
		log.Printf("DBKeeper.GetVisitInfo(2): select query error: %v\n", err)
		return DBResult{err, nil}
	}
	defer rows.Close()
	for rows.Next() {
		var tmp VisitRow
		err = rows.Scan(&tmp.Id, &tmp.Name, &tmp.Count, &tmp.Price,
			&tmp.PriceZnvlp, &tmp.Reason, &tmp.PayDt, &tmp.PrpCount,
			&tmp.PrevCountSaved, &tmp.PrevCount)
		if err != nil {
			log.Printf("DBKeeper.GetVisitInfo(2): get row error: %v\n", err)
			resultError = err
		}
		resultData.Rows = append(resultData.Rows, tmp)
	}
	///////
	return DBResult{resultError, resultData}
}
