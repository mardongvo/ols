package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	//"sort"
)

type ZnvlpPriceItem struct {
	Id      int     `json:"id"`
	IdZnvlp int     `json:"id_znvlp"`
	Price   float64 `json:"price"`
	Date    string  `json:"dt"`
}

type ZnvlpItem struct {
	Id         int     `json:"id"`
	Name       string  `json:"name"`
	MaxPrice   float64 `json:"price"`
	MaxDate    string  `json:"dt"`
	Similarity float64 `json:"similarity"`
}

type ZnvlpDictionaryItem struct {
	Id     int              `json:"id"`
	Name   string           `json:"name"`
	Prices []ZnvlpPriceItem `json:"prices"`
}

//
type RequestNewZnvlp struct {
	Name string `json:"name"`
}
type RequestSaveZnvlp struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type RequestDeleteZnvlp struct {
	Id int `json:"id"`
}

//
func JsonApiZnvlp(w http.ResponseWriter, r *http.Request) {
	var res DBResult
	if r.Method == "GET" {
		res = dbkeeper.GetZnvlpList()
	}
	if r.Method == "POST" {
		inp, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("JsonApiZnvlp(1):POST: error %v, source: %s", err, r.Body)
			res = DBResult{fmt.Errorf("%v", err), nil}
			goto END
		}
		var rq RequestNewZnvlp
		err = json.Unmarshal(inp, &rq)
		if err != nil {
			log.Printf("JsonApiZnvlp(2):POST: error %v", err)
			res = DBResult{fmt.Errorf("%v", err), nil}
			goto END
		}
		res = dbkeeper.AddZnvlp(rq)
	}
	if r.Method == "PUT" {
		inp, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("JsonApiZnvlp(1):PUT: error %v, source: %s", err, r.Body)
			res = DBResult{fmt.Errorf("%v", err), nil}
			goto END
		}
		var rq RequestSaveZnvlp
		err = json.Unmarshal(inp, &rq)
		if err != nil {
			log.Printf("JsonApiZnvlp(2):PUT: error %v", err)
			res = DBResult{fmt.Errorf("%v", err), nil}
			goto END
		}
		res = dbkeeper.SaveZnvlp(rq)
	}
	if r.Method == "DELETE" {
		inp, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("JsonApiZnvlp(1):DELETE: error %v, source: %s", err, r.Body)
			res = DBResult{fmt.Errorf("%v", err), nil}
			goto END
		}
		var rq RequestDeleteZnvlp
		err = json.Unmarshal(inp, &rq)
		if err != nil {
			log.Printf("JsonApiZnvlp(2):DELETE: error %v", err)
			res = DBResult{fmt.Errorf("%v", err), nil}
			goto END
		}
		res = dbkeeper.DeleteZnvlp(rq)
	}
END:
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiZnvlp: error %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Printf("JsonApiZnvlp: write error %v", err)
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
		rows, err = dk.db.Query(`select id, id_znvlp, price, dt from znvlp_maxprice where id_znvlp in (select id from znvlp) order by id_znvlp, dt;`)
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

func (dk *DBKeeper) AddZnvlp(rq RequestNewZnvlp) DBResult {
	tx, err := dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.AddZnvlp(1): begin tx error: %v\n", err)
		return DBResult{err, nil}
	}
	_, err = tx.Exec(`insert into znvlp(name) values($1);`, strings.ToUpper(rq.Name))
	if err != nil {
		tx.Rollback()
		log.Printf("DBKeeper.AddZnvlp(2): insert error: %v\n", err)
		return DBResult{err, nil}
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("DBKeeper.AddZnvlp(3): commit tx error %v", err)
	}
	return DBResult{err, nil}
}

func (dk *DBKeeper) SaveZnvlp(rq RequestSaveZnvlp) DBResult {
	tx, err := dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.SaveZnvlp(1): begin tx error: %v\n", err)
		return DBResult{err, nil}
	}
	_, err = tx.Exec(`update znvlp set name=$1 where id=$2;`, strings.ToUpper(rq.Name), rq.Id)
	if err != nil {
		tx.Rollback()
		log.Printf("DBKeeper.SaveZnvlp(2): insert error: %v\n", err)
		return DBResult{err, nil}
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("DBKeeper.SaveZnvlp(3): commit tx error %v", err)
	}
	return DBResult{err, nil}
}

func (dk *DBKeeper) DeleteZnvlp(rq RequestDeleteZnvlp) DBResult {
	tx, err := dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.DeleteZnvlp(1): begin tx error: %v\n", err)
		return DBResult{err, nil}
	}
	_, err = tx.Exec(`delete from znvlp where id=$1;`, rq.Id)
	if err != nil {
		tx.Rollback()
		log.Printf("DBKeeper.DeleteZnvlp(2): delete error: %v\n", err)
		return DBResult{err, nil}
	}
	_, err = tx.Exec(`delete from znvlp_maxprice where id_znvlp=$1;`, rq.Id)
	if err != nil {
		tx.Rollback()
		log.Printf("DBKeeper.DeleteZnvlp(2): delete error: %v\n", err)
		return DBResult{err, nil}
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("DBKeeper.DeleteZnvlp(3): commit tx error %v", err)
	}
	return DBResult{err, nil}
}

////////////////
type RequestNewPrice struct {
	IdZnvlp int     `json:"id_znvlp"`
	Price   float64 `json:"price"`
	Date    string  `json:"dt"`
}

type RequestSavePrice struct {
	Id    int     `json:"id"`
	Price float64 `json:"price"`
	Date  string  `json:"dt"`
}

type RequestDeletePrice struct {
	Id int `json:"id"`
}

func JsonApiZnvlpPrice(w http.ResponseWriter, r *http.Request) {
	var res DBResult
	if r.Method == "POST" {
		inp, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("JsonApiZnvlpPrice(1):POST: error %v, source: %s", err, r.Body)
			res = DBResult{fmt.Errorf("%v", err), nil}
			goto END
		}
		var rq RequestNewPrice
		err = json.Unmarshal(inp, &rq)
		if err != nil {
			log.Printf("JsonApiZnvlpPrice(2):POST: error %v", err)
			res = DBResult{fmt.Errorf("%v", err), nil}
			goto END
		}
		res = dbkeeper.AddPrice(rq)
	}
	if r.Method == "PUT" {
		inp, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("JsonApiZnvlpPrice(1):PUT: error %v, source: %s", err, r.Body)
			res = DBResult{fmt.Errorf("%v", err), nil}
			goto END
		}
		var rq RequestSavePrice
		err = json.Unmarshal(inp, &rq)
		if err != nil {
			log.Printf("JsonApiZnvlpPrice(2):PUT: error %v", err)
			res = DBResult{fmt.Errorf("%v", err), nil}
			goto END
		}
		res = dbkeeper.SavePrice(rq)
	}
	if r.Method == "DELETE" {
		inp, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("JsonApiZnvlpPrice(1):DELETE: error %v, source: %s", err, r.Body)
			res = DBResult{fmt.Errorf("%v", err), nil}
			goto END
		}
		var rq RequestDeletePrice
		err = json.Unmarshal(inp, &rq)
		if err != nil {
			log.Printf("JsonApiZnvlpPrice(2):DELETE: error %v", err)
			res = DBResult{fmt.Errorf("%v", err), nil}
			goto END
		}
		res = dbkeeper.DeletePrice(rq)
	}
END:
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiZnvlpPrice: error %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Printf("JsonApiZnvlpPrice: write error %v", err)
		return
	}
}

func (dk *DBKeeper) AddPrice(rq RequestNewPrice) DBResult {
	tx, err := dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.AddPrice(1): begin tx error: %v\n", err)
		return DBResult{err, nil}
	}
	_, err = tx.Exec(`insert into znvlp_maxprice(id_znvlp, dt, price) values($1,$2,$3);`, rq.IdZnvlp, rq.Date, rq.Price)
	if err != nil {
		tx.Rollback()
		log.Printf("DBKeeper.AddPrice(2): insert error: %v\n", err)
		return DBResult{err, nil}
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("DBKeeper.AddPrice(3): commit tx error %v", err)
	}
	return DBResult{err, nil}
}

func (dk *DBKeeper) SavePrice(rq RequestSavePrice) DBResult {
	tx, err := dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.SavePrice(1): begin tx error: %v\n", err)
		return DBResult{err, nil}
	}
	_, err = tx.Exec(`update znvlp_maxprice set dt=$1, price=$2 where id=$3;`, rq.Date, rq.Price, rq.Id)
	if err != nil {
		tx.Rollback()
		log.Printf("DBKeeper.SavePrice(2): insert error: %v\n", err)
		return DBResult{err, nil}
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("DBKeeper.SavePrice(3): commit tx error %v", err)
	}
	return DBResult{err, nil}
}

func (dk *DBKeeper) DeletePrice(rq RequestDeletePrice) DBResult {
	tx, err := dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.DeletePrice(1): begin tx error: %v\n", err)
		return DBResult{err, nil}
	}
	_, err = tx.Exec(`delete from znvlp_maxprice where id=$1;`, rq.Id)
	if err != nil {
		tx.Rollback()
		log.Printf("DBKeeper.DeletePrice(2): insert error: %v\n", err)
		return DBResult{err, nil}
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("DBKeeper.DeletePrice(3): commit tx error %v", err)
	}
	return DBResult{err, nil}
}
