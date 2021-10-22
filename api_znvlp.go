package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type ZnvlpPriceItem struct {
	Id      int     `json:"id"`
	IdZnvlp int     `json:"id_znvlp"`
	Price   float64 `json:"price"`
	Date    string  `json:"dt"`
}

type ZnvlpDictionaryItem struct {
	Id     int              `json:"id"`
	Name   string           `json:"name"`
	Prices []ZnvlpPriceItem `json:"prices"`
}

//
type CommonRequest interface {
	Action(tx *sql.Tx) error
}

type RequestConstructor func(inp []byte) (CommonRequest, error)

var ACTION_MAP = []struct {
	path       string
	method     string
	constuctor RequestConstructor
}{
	{
		"/api/znvlp",
		"POST",
		func(inp []byte) (CommonRequest, error) {
			request := RequestNewZnvlp{}
			err := json.Unmarshal(inp, &request)
			return request, err
		},
	},
	{
		"/api/znvlp",
		"PUT",
		func(inp []byte) (CommonRequest, error) {
			request := RequestSaveZnvlp{}
			err := json.Unmarshal(inp, &request)
			return request, err
		},
	},
	{
		"/api/znvlp",
		"DELETE",
		func(inp []byte) (CommonRequest, error) {
			request := RequestDeleteZnvlp{}
			err := json.Unmarshal(inp, &request)
			return request, err
		},
	},
	{
		"/api/znvlp_price",
		"POST",
		func(inp []byte) (CommonRequest, error) {
			request := RequestNewPrice{}
			err := json.Unmarshal(inp, &request)
			return request, err
		},
	},
	{
		"/api/znvlp_price",
		"PUT",
		func(inp []byte) (CommonRequest, error) {
			request := RequestSavePrice{}
			err := json.Unmarshal(inp, &request)
			return request, err
		},
	},
	{
		"/api/znvlp_price",
		"DELETE",
		func(inp []byte) (CommonRequest, error) {
			request := RequestDeletePrice{}
			err := json.Unmarshal(inp, &request)
			return request, err
		},
	},
}

type RequestNewZnvlp struct {
	Name string `json:"name"`
}

func (rq RequestNewZnvlp) Action(tx *sql.Tx) error {
	_, err := tx.Exec(`insert into znvlp(name) values($1);`, strings.ToUpper(rq.Name))
	return err
}

type RequestSaveZnvlp struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (rq RequestSaveZnvlp) Action(tx *sql.Tx) error {
	_, err := tx.Exec(`update znvlp set name=$1 where id=$2;`, strings.ToUpper(rq.Name), rq.Id)
	return err
}

type RequestDeleteZnvlp struct {
	Id int `json:"id"`
}

func (rq RequestDeleteZnvlp) Action(tx *sql.Tx) error {
	_, err := tx.Exec(`delete from znvlp where id=$1;`, rq.Id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`delete from znvlp_maxprice where id_znvlp=$1;`, rq.Id)
	if err != nil {
		return err
	}
	return nil
}

type RequestNewPrice struct {
	IdZnvlp int     `json:"id_znvlp"`
	Price   float64 `json:"price"`
	Date    string  `json:"dt"`
}

func (rq RequestNewPrice) Action(tx *sql.Tx) error {
	_, err := tx.Exec(`insert into znvlp_maxprice(id_znvlp, dt, price) values($1,$2,$3);`, rq.IdZnvlp, rq.Date, rq.Price)
	return err
}

type RequestSavePrice struct {
	Id    int     `json:"id"`
	Price float64 `json:"price"`
	Date  string  `json:"dt"`
}

func (rq RequestSavePrice) Action(tx *sql.Tx) error {
	_, err := tx.Exec(`update znvlp_maxprice set dt=$1, price=$2 where id=$3;`, rq.Date, rq.Price, rq.Id)
	return err
}

type RequestDeletePrice struct {
	Id int `json:"id"`
}

func (rq RequestDeletePrice) Action(tx *sql.Tx) error {
	_, err := tx.Exec(`delete from znvlp_maxprice where id=$1;`, rq.Id)
	return err
}

//
func JsonApiZnvlp(w http.ResponseWriter, r *http.Request) {
	var res DBResult
	var mapped bool = false
	if (r.URL.Path == "/api/znvlp") && (r.Method == "GET") {
		res = dbkeeper.GetZnvlpList()
		goto END
	}
	for _, act := range ACTION_MAP {
		if (r.URL.Path == act.path) && (r.Method == act.method) {
			mapped = true
			inp, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("JsonApiZnvlp:(%s, %s): read error %v, source: %s", act.path, act.method, err, r.Body)
				res = DBResult{fmt.Errorf("%v", err), nil}
				goto END
			}
			request, err := act.constuctor(inp)
			if err != nil {
				log.Printf("JsonApiZnvlp:(%s, %s): unmarshal error %v", act.path, act.method, err)
				res = DBResult{fmt.Errorf("%v", err), nil}
				goto END
			}
			res = dbkeeper.CommonAction(request)
		}
	}
	if !mapped {
		log.Printf("JsonApiZnvlp:(%s, %s): unmapped request", r.URL.Path, r.Method)
		res = DBResult{fmt.Errorf("Unmapped request"), nil}
		goto END
	}
END:
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiZnvlp:(%s,%s): marshal error %v", r.URL.Path, r.Method, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Printf("JsonApiZnvlp:(%s,%s): write error %v", r.URL.Path, r.Method, err)
		return
	}
}

func (dk *DBKeeper) GetZnvlpList() DBResult {
	var resultData []ZnvlpDictionaryItem
	resultData = make([]ZnvlpDictionaryItem, 0)
	//1. names
	rows, err := dk.db.Query(`select id, upper(name) from znvlp order by upper(name);`)
	if err != nil {
		log.Printf("DBKeeper.GetZnvlpList(1): select query error: %v\n", err)
		return DBResult{err, nil}
	}
	for rows.Next() {
		var tmp ZnvlpDictionaryItem
		tmp.Prices = make([]ZnvlpPriceItem, 0)
		err = rows.Scan(&tmp.Id, &tmp.Name)
		if err != nil {
			log.Printf("DBKeeper.GetZnvlpList(1): get row error: %v\n", err)
		}
		resultData = append(resultData, tmp)
	}
	rows.Close()
	//2. prices
	if len(resultData) > 0 {
		rows, err = dk.db.Query(`select id, id_znvlp, price, dt from znvlp_maxprice where id_znvlp in (select id from znvlp) order by id_znvlp, dt desc;`)
		if err != nil {
			log.Printf("DBKeeper.GetZnvlpList(2): select query error: %v\n", err)
			return DBResult{err, nil}
		}
		lastZnum := 0
		for rows.Next() {
			var tmp ZnvlpPriceItem
			err = rows.Scan(&tmp.Id, &tmp.IdZnvlp, &tmp.Price, &tmp.Date)
			tmp.Date = tmp.Date[0:10] //only date without time
			if err != nil {
				log.Printf("DBKeeper.GetZnvlpList(2): get row error: %v\n", err)
			}
			if resultData[lastZnum].Id != tmp.IdZnvlp {
				for i, _ := range resultData {
					if resultData[i].Id == tmp.IdZnvlp {
						lastZnum = i
						break
					}
				}
			}
			resultData[lastZnum].Prices = append(resultData[lastZnum].Prices, tmp)
		}
	}
	//
	return DBResult{nil, resultData}
}

func (dk *DBKeeper) CommonAction(rq CommonRequest) DBResult {
	tx, err := dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.CommonAction(1): begin tx error: %v, request: %#v\n", err, rq)
		return DBResult{err, nil}
	}
	err = rq.Action(tx)
	if err != nil {
		tx.Rollback()
		log.Printf("DBKeeper.CommonAction(2): action error: %v, request: %#v\n", err, rq)
		return DBResult{err, nil}
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("DBKeeper.CommonAction(2): commit tx error: %v, request: %#v\n", err, rq)
	}
	return DBResult{err, nil}
}
