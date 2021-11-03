package main

import (
	"log"
	"sort"
)

type ZnvlpHintItem struct {
	Id         int     `json:"id"`
	Name       string  `json:"name"`
	MaxPrice   float64 `json:"price"`
	MaxDate    string  `json:"dt"`
	Similarity float64 `json:"similarity"`
}

type ZnvlpHints []ZnvlpHintItem

func (a ZnvlpHints) Len() int           { return len(a) }
func (a ZnvlpHints) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ZnvlpHints) Less(i, j int) bool { return a[i].Similarity > a[j].Similarity }

func (dk *DBKeeper) GetZnvlpHints() ZnvlpHints {
	var resultData ZnvlpHints
	resultData = make(ZnvlpHints, 0)
	rows, err := dk.db.Query(`with maxp as (select distinct on (id_znvlp)
	   id_znvlp, dt, price from znvlp_maxprice order by id_znvlp, dt desc, price)
	select z.id, z.name, maxp.price, maxp.dt from znvlp z, maxp where
	   z.id=maxp.id_znvlp order by z.name;`)
	if err != nil {
		log.Printf("DBKeeper.GetZnvlpHint(1): select query error: %v\n", err)
		return resultData
	}
	defer rows.Close()
	for rows.Next() {
		var tmp ZnvlpHintItem
		err = rows.Scan(&tmp.Id, &tmp.Name, &tmp.MaxPrice, &tmp.MaxDate)
		tmp.MaxDate = tmp.MaxDate[0:10]
		if err != nil {
			log.Printf("DBKeeper.GetZnvlpHint(1): get row error: %v\n", err)
		}
		resultData = append(resultData, tmp)
	}
	return resultData
}

//return
func SimilarityCopy(hints ZnvlpHints, target string, minimalSimilarity float64) ZnvlpHints {
	var res ZnvlpHints = ZnvlpHints{ZnvlpHintItem{Id: 0, Name: "-", MaxDate: "",
		MaxPrice: 0, Similarity: minimalSimilarity}}
	for _, h := range hints {
		newh := h
		newh.Similarity = similarity2(h.Name, target)
		res = append(res, newh)
	}
	sort.Sort(res)
	return res
}
