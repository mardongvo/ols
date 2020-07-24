package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/tealeg/xlsx"
)

func ApiGetXls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	var idd int64
	var err error
	var f *xlsx.File
	var cookie *http.Cookie
	idd, err = strconv.ParseInt(r.URL.Query().Get("id"), 10, 32)
	if err != nil {
		idd = 0
	}
	data, rows := dbkeeper.GetXlsData(int(idd))
	cookie, err = r.Cookie("worker")
	if err != nil {
		data.Worker = "________"
	} else {
		data.Worker, _ = url.QueryUnescape(cookie.Value)
	}
	//
	f, err = XlsRenderTemplate("./template/visit_template.xlsx", data, rows)
	if err != nil {
		log.Printf("ApiGetXls: error %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Xlsx template rendering error"))
		return
	}
	err = f.Write(w)
	if err != nil {
		log.Printf("ApiGetXls: write error %v", err)
		return
	}
}

func (dk *DBKeeper) GetXlsData(id int) (XlsData, []XlsRow) {
	var resultData XlsData
	var resultRows []XlsRow
	var n int
	resultRows = make([]XlsRow, 0)
	/////////
	rows, err := dk.db.Query(`select prp.num, prp.dtbeg, prp.dtend, person.fio
	 from person, prp, visit where visit.id=$1
	 and visit.id_prp=prp.id and visit.id_own=person.id;`, id)
	if err != nil {
		log.Printf("DBKeeper.GetXlsData(1): select query error: %v\n", err)
		return resultData, resultRows
	}
	defer rows.Close()
	for rows.Next() {
		var dt1, dt2 time.Time
		err = rows.Scan(&resultData.PrpNum, &dt1, &dt2, &resultData.PersonFio)
		if err != nil {
			log.Printf("DBKeeper.GetXlsData(1): get row error: %v\n", err)
		}
		resultData.PrpDtBeg = fmt.Sprintf("%02d.%02d.%04d", dt1.Day(), dt1.Month(), dt1.Year())
		resultData.PrpDtEnd = fmt.Sprintf("%02d.%02d.%04d", dt2.Day(), dt2.Month(), dt2.Year())
	}
	///////
	rows, err = dk.db.Query(`with
		curinfo as (select * from visit_info where id_own = $1),
		prev_visits as (select a.id id from visit a, visit b where b.id=$1 and b.id_prp=a.id_prp and a.dt<b.dt),
		prev_counts as (select id_prpt, sum(cnt) cnt from visit_info where id_own in (select id from prev_visits) group by id_prpt)
	select pt.name, curinfo.cnt, curinfo.price, curinfo.price_znvlp,
    	curinfo.reason, curinfo.paydt, pt.cnt prp_count,
    	coalesce(prev_counts.cnt,0) prevcnt
		from prp_template pt 
        inner join curinfo on (pt.id = curinfo.id_prpt) 
        left outer join prev_counts on (pt.id = prev_counts.id_prpt)
        order by lower(pt.name);`, id)
	if err != nil {
		log.Printf("DBKeeper.GetXlsData(2): select query error: %v\n", err)
		return resultData, resultRows
	}
	defer rows.Close()
	for rows.Next() {
		var tmp XlsRow
		n++
		tmp.N = n
		err = rows.Scan(&tmp.Name, &tmp.Count, &tmp.Price, &tmp.PriceZnvlp,
			&tmp.Reason, &tmp.PayDt, &tmp.PrpCount,
			&tmp.PrevCount)
		if err != nil {
			log.Printf("DBKeeper.GetXlsData(2): get row error: %v\n", err)
		}
		//**** вычисления количества и суммы к оплате; взято из visit.tmpl
		limit := tmp.PrpCount - tmp.PrevCount
		if limit < 0 {
			limit = 0
		}
		//количество к оплате
		tmp.PayCount = tmp.Count
		if limit < tmp.Count {
			tmp.PayCount = limit
		}
		//сумма к оплате
		tmp.Pay = tmp.Price
		if tmp.Count > 0 {
			if (tmp.PriceZnvlp > 0) && (tmp.PriceZnvlp < (tmp.Price / float64(tmp.Count))) {
				tmp.Pay = math.Round(tmp.PriceZnvlp*float64(tmp.PayCount)*100) / 100
			} else {
				tmp.Pay = math.Round(tmp.Price*float64(tmp.PayCount)/float64(tmp.Count)*100) / 100
			}
		}
		if tmp.Count == 0 {
			tmp.Pay = 0
		}
		tmp.NotPay = tmp.Price - tmp.Pay
		tmp.Remain = tmp.PrpCount - tmp.PrevCount - tmp.PayCount
		//****
		//
		resultData.SumPrice += tmp.Price
		resultData.SumPay += tmp.Pay
		resultData.SumNotPay += tmp.NotPay
		resultRows = append(resultRows, tmp)
	}
	///////
	return resultData, resultRows
}
