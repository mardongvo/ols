package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

type FarmItem struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func JsonApiFarmList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var res DBResult
	res = dbkeeper.GetFarmList()
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiFarmList: error %v", err)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		log.Printf("JsonApiFarmList: write error %v", err)
		return
	}
}

func (dk *DBKeeper) GetFarmList() DBResult {
	var resultData []FarmItem
	resultData = make([]FarmItem, 0)
	rows, err := dk.db.Query(`select id, upper(name) from farma order by upper(name);`)
	if err != nil {
		log.Printf("DBKeeper.GetFarmList(1): select query error: %v\n", err)
		return DBResult{err, nil}
	}
	defer rows.Close()
	for rows.Next() {
		var tmp FarmItem
		err = rows.Scan(&tmp.Id, &tmp.Name)
		if err != nil {
			log.Printf("DBKeeper.GetFarmList(1): get row error: %v\n", err)
		}
		resultData = append(resultData, tmp)
	}
	return DBResult{nil, resultData}

}

//////////

type TemplateCadidate struct {
	Id             int         `json:"id"`
	Name           string      `json:"name"`
	IdPrp          int         `json:"id_prp"`
	FarmCandidates FCandidates `json:"candidates"`
}

type FarmCandidate struct {
	Id           int     `json:"id"`
	Name         string  `json:"name"`
	MatchPercent float64 `json:"match"`
}

type FCandidates []FarmCandidate

func (a FCandidates) Len() int           { return len(a) }
func (a FCandidates) Less(i, j int) bool { return a[i].MatchPercent > a[j].MatchPercent }
func (a FCandidates) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func JsonApiFarmCandidates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var res DBResult
	res = dbkeeper.GetTemplateCandidates(20, 0.3) //TODO: param,config?
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiFarmCandidates: error %v", err)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		log.Printf("JsonApiFarmCandidates: write error %v", err)
		return
	}
}

func (dk *DBKeeper) GetTemplateCandidates(limit int, addPercent float64) DBResult {
	var tc []TemplateCadidate
	var rows *sql.Rows
	var err error
	tc = make([]TemplateCadidate, 0)
	rows, err = dk.db.Query(`select id, name, id_own from prp_template where name<>'??' and name>''
		and id_farm=0 and id_own in (select id from prp where active=1) order by name limit $1;`, limit)
	if err != nil {
		log.Printf("DBKeeper.GetTemplateCandidates(1): select query error: %v\n", err)
		return DBResult{err, nil}
	}
	for rows.Next() {
		var tmp TemplateCadidate
		tmp.FarmCandidates = make(FCandidates, 0)
		err = rows.Scan(&tmp.Id, &tmp.Name, &tmp.IdPrp)
		if err != nil {
			log.Printf("DBKeeper.GetTemplateCandidates(2): get row error: %v\n", err)
		} else {
			tc = append(tc, tmp)
		}
	}
	rows.Close()
	//candidates from farma
	for i, v := range tc {
		for step := 1; step < 3; step++ {
			if step == 1 { //кандидаты из справочника лекарств
				rows, err = dk.db.Query(`select id, name, similarity($1,name) from farma where id>0
			and ($1 % name) and similarity($1,name)>=$2;`, v.Name, addPercent)
			}
			if step == 2 { //кандидаты по совпадению из ПРП
				rows, err = dk.db.Query(`select f.id, f.name, max(similarity($1,pt.name))
				from farma f, prp_template pt where f.id>0 and f.id=pt.id_farm
				and ($1 % pt.name) and similarity($1,pt.name)>=$2 group by f.id, f.name;`, v.Name, addPercent)
			}
			if err != nil {
				log.Printf("DBKeeper.GetTemplateCandidates(3): select query error: %v\n", err)
				continue
			}
			needMore := true
			for rows.Next() {
				var tmp FarmCandidate
				err = rows.Scan(&tmp.Id, &tmp.Name, &tmp.MatchPercent)
				if err != nil {
					log.Printf("DBKeeper.GetTemplateCandidates(4): get row error: %v\n", err)
				} else {
					tc[i].FarmCandidates = append(tc[i].FarmCandidates, tmp)
					if tmp.MatchPercent >= 0.9 { //TODO: config?
						needMore = false
					}
				}
			}
			rows.Close()
			if !needMore {
				break
			}
		} //step
		sort.Sort(tc[i].FarmCandidates)
	}
	return DBResult{nil, tc}
}

////////////////////////

type SaveTemplateRequest struct {
	IdTemplate int `json:"id_template"`
	IdFarm     int `json:"id_farm"`
}

func JsonApiCandidatesSave(w http.ResponseWriter, r *http.Request) {
	var res DBResult
	var rq []SaveTemplateRequest
	inp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("JsonApiCandidatesSave(1): error %v", err)
		res = DBResult{fmt.Errorf("%v", err), nil}
		goto END
	}
	err = json.Unmarshal(inp, &rq)
	if err != nil {
		log.Printf("JsonApiCandidatesSave(2): error %v", err)
		res = DBResult{fmt.Errorf("%v", err), nil}
		goto END
	}
	err = dbkeeper.SaveTemplates(rq)
	if err != nil {
		log.Printf("JsonApiCandidatesSave(3): error %v", err)
		res = DBResult{fmt.Errorf("%v", err), nil}
		goto END
	}
	res = DBResult{nil, ""}
END:
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiCandidatesSave(4): error %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Printf("JsonApiCandidatesSave: write error %v", err)
		return
	}
}

func (dk *DBKeeper) SaveTemplates(rq []SaveTemplateRequest) error {
	tx, err := dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.SaveTemplates(1): begin tx error: %v\n", err)
		return err
	}
	for _, v := range rq {
		_, err := tx.Exec(`update prp_template set id_farm=$1 where id=$2;`,
			v.IdFarm, v.IdTemplate)
		if err != nil {
			tx.Rollback()
			log.Printf("DBKeeper.SaveTemplates(2): update error: %v\n", err)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("DBKeeper.SaveTemplates(3): commit tx error %v", err)
	}
	return err
}

////////////////////

func JsonApiFarmAddNew(w http.ResponseWriter, r *http.Request) {
	var res DBResult
	w.Header().Set("Content-Type", "application/json")
	newname := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("name")))
	if newname > "" {
		res = DBResult{dbkeeper.FarmAddNew(newname), ""}
	} else {
		res = DBResult{nil, ""}
	}
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiFarmAddNew: error %v", err)
		return
	}
	w.Write(data)
}

func (dk *DBKeeper) FarmAddNew(name string) error {
	tx, err := dk.db.Begin()
	if err != nil {
		log.Printf("DBKeeper.FarmAddNew(1): begin tx error: %v\n", err)
		return err
	}
	_, err = tx.Exec(`insert into farma(name) select $1::varchar from farma where not exists
	 (select * from farma where similarity($1::varchar, name)=1.0) limit 1;`,
		name)
	if err != nil {
		tx.Rollback()
		log.Printf("DBKeeper.FarmAddNew(2): insert error: %v\n", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("DBKeeper.FarmAddNew(3): commit tx error %v", err)
	}
	return err
}
