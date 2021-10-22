package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const DateLayout = "2006-01-02"

type VisitRow struct {
	Id             int        `json:"id"` //prp_template id
	Name           string     `json:"name"`
	Count          int        `json:"count"`
	Price          float32    `json:"price"`
	PriceZnvlp     float32    `json:"price_znvlp"`
	Reason         string     `json:"reason"`
	PayDt          string     `json:"paydt"`
	PrpCount       int        `json:"prp_count"`
	PrevCount      int        `json:"prev_count"`       //
	PrevCountSaved int        `json:"prev_count_saved"` //
	RecepCount     int        `json:"recep_count"`
	Hints          ZnvlpHints `json:"hints"`
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

type VisitAddResponse struct {
	Id int `json:"id"`
}

type VisitRemoveRequest struct {
	Id int `json:"id"`
}

func JsonApiVisitInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var res DBResult
	var tmp int
	var dt string
	var idPrp, idd int64
	var err error
	idd, err = strconv.ParseInt(r.URL.Query().Get("id"), 10, 32)
	if err != nil {
		idd = 0
	}
	dt = r.URL.Query().Get("dt")
	idPrp, err = strconv.ParseInt(r.URL.Query().Get("id_prp"), 10, 32)
	if err != nil {
		idPrp = 0
	}
	if (idPrp > 0) && (dt > "") {
		tmp, err = dbkeeper.AddVisit(int(idPrp), dt)
		if err != nil {
			res = DBResult{err, nil}
		} else {
			res = DBResult{nil, VisitAddResponse{tmp}}
		}
	} else {
		res = dbkeeper.GetVisitInfo(int(idd))
	}
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiVisitInfo: error %v", err)
		return
	}
	w.Write(data)
}

func JsonApiVisitRemove(w http.ResponseWriter, r *http.Request) {
	var res DBResult
	var rq VisitRemoveRequest
	inp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("JsonApiVisitRemove(1): error %v", err)
		res = DBResult{fmt.Errorf("%v", err), nil}
		goto END
	}
	err = json.Unmarshal(inp, &rq)
	if err != nil {
		log.Printf("JsonApiVisitRemove(2): error %v", err)
		res = DBResult{fmt.Errorf("%v", err), nil}
		goto END
	}
	err = dbkeeper.RemoveVisitInfo(rq.Id, true)
	if err != nil {
		log.Printf("JsonApiVisitRemove(3): error %v", err)
		res = DBResult{fmt.Errorf("%v", err), nil}
		goto END
	}
	res = DBResult{nil, nil}
END:
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiVisitRemove(4): error %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Printf("JsonApiVisitRemove: write error %v", err)
		return
	}
}

func JsonApiVisitSave(w http.ResponseWriter, r *http.Request) {
	var res DBResult
	var rq VisitSaveRequest
	inp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("JsonApiVisitSave(1): error %v", err)
		res = DBResult{fmt.Errorf("%v", err), nil}
		goto END
	}
	err = json.Unmarshal(inp, &rq)
	if err != nil {
		log.Printf("JsonApiVisitSave(2): error %v", err)
		res = DBResult{fmt.Errorf("%v", err), nil}
		goto END
	}
	err = dbkeeper.SaveVisitInfo(rq)
	if err != nil {
		log.Printf("JsonApiVisitSave(3): error %v", err)
		res = DBResult{fmt.Errorf("%v", err), nil}
		goto END
	}
	res = dbkeeper.GetVisitInfo(rq.Id)
END:
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiVisitSave(4): error %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Printf("JsonApiVisitSave: write error %v", err)
		return
	}
}

func (dk *DBKeeper) GetVisitInfo(id int) DBResult {
	var resultError error = nil
	var resultData VisitInfo
	//
	hints := dk.GetZnvlpHints()
	//
	resultData.Rows = make([]VisitRow, 0)
	//
	rows, err := dk.db.Query(`select visit.id, visit.dt,
	 prp.id, prp.num, prp.dtbeg, prp.dtend,
	 person.id, person.fio, person.ndoc
	 from person, prp, visit where visit.id=$1
	 and visit.id_prp=prp.id and visit.id_own=person.id;`, id)
	if err != nil {
		log.Printf("DBKeeper.GetVisitInfo(1): select query error: %v\n", err)
		return DBResult{err, nil}
	}
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
	rows.Close()
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
		tmp.Hints = SimilarityCopy(hints, tmp.Name, 0.8)
		if err != nil {
			log.Printf("DBKeeper.GetVisitInfo(2): get row error: %v\n", err)
			resultError = err
		}
		resultData.Rows = append(resultData.Rows, tmp)
	}
	///////
	return DBResult{resultError, resultData}
}

func (dk *DBKeeper) SaveVisitInfo(rq VisitSaveRequest) error {
	tx, err := dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.SaveVisitInfo(1): begin tx error: %v\n", err)
		return err
	}
	for _, v := range rq.Rows {
		_, err := tx.Exec(`update visit_info set cnt=$1, price=$2, price_znvlp=$3,
			reason=$4::varchar, paydt=$5::varchar where id_prpt=$6 and id_own=$7;`,
			v.Count, v.Price, v.PriceZnvlp, v.Reason, v.PayDt, v.Id, rq.Id)
		if err != nil {
			tx.Rollback()
			log.Printf("DBKeeper.SaveVisitInfo(2): update error: %v\n", err)
			return err
		}
	}
	return tx.Commit()
}

func (dk *DBKeeper) RemoveVisitInfo(id int, onlyZero bool) error {
	var tx *sql.Tx
	var err error
	var rows *sql.Rows
	var doRemove bool = true
	tx, err = dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.RemoveVisitInfo(1): begin tx error: %v\n", err)
		return err
	}
	rows, err = tx.Query(`select coalesce(sum(cnt),0), coalesce(sum(price), 0.0), coalesce(sum(price_znvlp), 0.0),
		coalesce(max(reason),''), coalesce(max(paydt), '') from visit_info where id_own=$1;`, id)
	if err != nil {
		log.Printf("DBKeeper.RemoveVisitInfo(2): query error: %v\n", err)
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cnt int
		var price, pricez float64
		var reason, paydt string
		err = rows.Scan(&cnt, &price, &pricez, &reason, &paydt)
		if err != nil {
			tx.Rollback()
			log.Printf("DBKeeper.RemoveVisitInfo(3): scan error: %v\n", err)
			return err
		}
		if cnt > 0 || price > 0 || pricez > 0 || reason > "" || paydt > "" {
			doRemove = false
		}
	}
	if doRemove {
		_, err = tx.Exec("delete from visit_info where id_own=$1;", id)
		if err != nil {
			tx.Rollback()
			log.Printf("DBKeeper.RemoveVisitInfo(4): delete error: %v\n", err)
			return err
		}
		_, err = tx.Exec("delete from visit where id=$1;", id)
		if err != nil {
			tx.Rollback()
			log.Printf("DBKeeper.RemoveVisitInfo(5): delete error: %v\n", err)
			return err
		}
	}
	return tx.Commit()
}

func (dk *DBKeeper) AddVisit(idPrp int, dt string) (int, error) {
	var err error
	var tx *sql.Tx
	var row *sql.Row
	var visitId int
	var dtParsed time.Time
	//проверяем строку даты на соответствие формату YYYY-MM-DD
	dtParsed, err = time.Parse(DateLayout, dt)
	if err != nil {
		log.Printf("DBKeeper.AddVisit(0): date parse error: %v\n", err)
		return 0, err
	}
	//открываем транзакцию
	tx, err = dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.AddVisit(1): begin tx error: %v\n", err)
		return 0, err
	}
	//добавялем визит на дату, если его не существует
	_, err = tx.Exec(`insert into visit(id_prp, id_own, dt)
		select $1, id_own, $2 from prp where id=$1 and not exists (
			select id from visit where id_prp=$1 and dt=$2
		);`, idPrp, dtParsed)
	if err != nil {
		log.Printf("DBKeeper.AddVisit(2): error: %v\n", err)
		tx.Rollback()
		return 0, err
	}
	//получаем Id визита
	row = tx.QueryRow("select id from visit where id_prp=$1 and dt=$2;",
		idPrp, dtParsed)
	err = row.Scan(&visitId)
	if err != nil {
		log.Printf("DBKeeper.AddVisit(3): error: %v\n", err)
		tx.Rollback()
		return 0, err
	}
	if visitId == 0 {
		tx.Rollback()
		return 0, fmt.Errorf("Ошибка добавления визита: visitId = 0")
	}
	//добавляем строки из шаблона ПРП только если их еще нет
	_, err = tx.Exec(`
	with c as (select id_prpt, sum(cnt) cnt from visit_info where
	 id_own in (select id from visit where id_prp=$1 and dt<$2) group by id_prpt)
	insert into visit_info(id_own, id_prpt, paydt, cnt, price, price_znvlp,
						reason, prevcnt)
		select $3, id, '', 0, 0, 0, '', coalesce(c.cnt, 0) from prp_template
            left outer join c on (c.id_prpt=prp_template.id)
        where prp_template.id_own=$1 and
            not exists (select 1 from visit_info where visit_info.id_own=$3
            and visit_info.id_prpt=prp_template.id);
	`, idPrp, dtParsed, visitId)
	if err != nil {
		log.Printf("DBKeeper.AddVisit(4): error: %v\n", err)
		tx.Rollback()
		return 0, err
	}
	err = tx.Commit()
	return visitId, err
}
