package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
)

type StatItem struct {
	PersonId      int     `json:"person_id"`
	PersonFio     string  `json:"person_fio"`
	PersonNdoc    string  `json:"person_ndoc"`
	PrpId         int     `json:"prp_id"`
	PrpNum        string  `json:"prp_num"`
	PrpDtBeg      string  `json:"prp_dtbeg"`
	PrpDtEnd      string  `json:"prp_dtend"`
	StatLastVisit string  `json:"stat_last_visit"`
	StatAllSum    float32 `json:"stat_all_sum"`
	StatPaySum    float32 `json:"stat_pay_sum"`
	StatExpensive int     `json:"stat_expensive"`
}

func JsonApiStatInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var res DBResult
	res = dbkeeper.GetStatInfo()
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JsonApiStatInfo: error %v", err)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		log.Printf("JsonApiStatInfo: write error %v", err)
		return
	}
}

func (dk *DBKeeper) GetStatInfo() DBResult {
	var resultData []StatItem
	resultData = make([]StatItem, 0)
	rows, err := dk.db.Query(`
	with
		-- средние цены
		tprice as (
		 select b.id_farm id_farm, sum(a.price)/sum(a.cnt) avgprice
		 from visit_info a, prp_template b
		 where a.id_prpt=b.id and a.cnt>0
		 group by b.id_farm
		),
		--остатки в сумме по прп
		tpaysum as (
		 select b.id_own id_prp, sum(COALESCE(a.cnt,0)*tprice.avgprice) pay_sum
		 from tprice, prp_template b left outer join visit_info a on a.id_prpt=b.id
		 where b.mark_del=0 and b.cnt<720 and
		 tprice.id_farm=b.id_farm
		 group by  b.id_own
		),
		-- сумма ПРП
		tallsum as (
		 select b.id_own id_prp, sum(COALESCE(b.cnt,0) * tprice.avgprice) sm
		 from prp_template b, tprice
		 where b.mark_del=0 and tprice.id_farm=b.id_farm
		 group by b.id_own
		),
		-- последний визит
		tlastvis as (
		 select id_prp, max(dt) dt from visit group by id_prp
		),
		-- дорогие лекарства
		texpensive as (
		 select id_own id_prp from prp_template where
		 upper(name) like '%ГИАЛГАН%' or upper(name) like '%БЕРОДУАЛ%' or upper(name) like '%СПИРИВА%'
		 group by id_own
		)
	select person.id, person.fio, person.ndoc,
	prp.id, prp.num, coalesce(to_char(prp.dtbeg, 'DD.MM.YYYY'), ''),
	coalesce(to_char(prp.dtend, 'DD.MM.YYYY'), ''),
	coalesce(to_char(tlastvis.dt, 'DD.MM.YYYY'), ''), tallsum.sm, tpaysum.pay_sum,
	case when prp.id in (select id_prp from texpensive) then 1 else 0 end
	from tallsum left outer join tlastvis on tallsum.id_prp=tlastvis.id_prp
	left outer join tpaysum on tallsum.id_prp=tpaysum.id_prp,
	person, prp where tallsum.id_prp=prp.id and person.id=prp.id_own
	and person.active=1 and prp.active=1
	order by person.fio;
	`)
	if err != nil {
		log.Printf("DBKeeper.GetStatInfo(1): select query error: %v\n", err)
		return DBResult{err, nil}
	}
	defer rows.Close()
	for rows.Next() {
		var tmp StatItem
		err = rows.Scan(&tmp.PersonId, &tmp.PersonFio, &tmp.PersonNdoc,
			&tmp.PrpId, &tmp.PrpNum, &tmp.PrpDtBeg, &tmp.PrpDtEnd,
			&tmp.StatLastVisit, &tmp.StatAllSum, &tmp.StatPaySum,
			&tmp.StatExpensive)
		if err != nil {
			log.Printf("DBKeeper.GetStatInfo(1): get row error: %v\n", err)
		}
		tmp.StatAllSum = float32(math.Round(float64(tmp.StatAllSum)*100) / 100.0)
		tmp.StatPaySum = float32(math.Round(float64(tmp.StatPaySum)*100) / 100.0)
		resultData = append(resultData, tmp)
	}
	return DBResult{nil, resultData}

}
