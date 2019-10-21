package main

import (
	"fmt"
	"strings"

	"github.com/tealeg/xlsx"
)

type XlsRow struct {
	N          int
	Name       string
	PrpCount   int
	PrevCount  int
	PayDt      string
	Count      int
	Remain     int
	PayCount   int
	Price      float64
	PriceZnvlp float64
	Pay        float64
	NotPay     float64
	Reason     string
}

type XlsData struct {
	PersonFio string
	PrpNum    string
	PrpDtBeg  string
	PrpDtEnd  string
	SumPrice  float64
	SumPay    float64
	SumNotPay float64
	Worker    string
}

func isIterableRow(row *xlsx.Row) bool {
	for _, cell := range row.Cells {
		if strings.Contains(cell.Value, "{{Row.") {
			return true
		}
	}
	return false
}

func xlsRenderCell(cell *xlsx.Cell, data XlsData, row XlsRow) {
	//int/float
	if cell.Value == "{{Row.N}}" {
		cell.SetInt(row.N)
		return
	}
	if cell.Value == "{{Row.PrpCount}}" {
		cell.SetInt(row.PrpCount)
		return
	}
	if cell.Value == "{{Row.PrevCount}}" {
		cell.SetInt(row.PrevCount)
		return
	}
	if cell.Value == "{{Row.Count}}" {
		cell.SetInt(row.Count)
		return
	}
	if cell.Value == "{{Row.Remain}}" {
		cell.SetInt(row.Remain)
		return
	}
	if cell.Value == "{{Row.PayCount}}" {
		cell.SetInt(row.PayCount)
		return
	}
	if cell.Value == "{{Row.Price}}" {
		cell.SetFloat(row.Price)
		cell.NumFmt = "0.00"
		return
	}
	if cell.Value == "{{Row.PriceZnvlp}}" {
		cell.SetFloat(row.PriceZnvlp)
		cell.NumFmt = "0.00"
		return
	}
	if cell.Value == "{{Row.Pay}}" {
		cell.SetFloat(row.Pay)
		cell.NumFmt = "0.00"
		return
	}
	if cell.Value == "{{Row.NotPay}}" {
		cell.SetFloat(row.NotPay)
		cell.NumFmt = "0.00"
		return
	}
	if cell.Value == "{{SumPrice}}" {
		cell.SetFloat(data.SumPrice)
		cell.NumFmt = "0.00"
		return
	}
	if cell.Value == "{{SumPay}}" {
		cell.SetFloat(data.SumPay)
		cell.NumFmt = "0.00"
		return
	}
	if cell.Value == "{{SumNotPay}}" {
		cell.SetFloat(data.SumNotPay)
		cell.NumFmt = "0.00"
		return
	}
	//string values
	res := cell.Value //для накопления замен
	res = strings.ReplaceAll(res, "{{PersonFio}}", data.PersonFio)
	res = strings.ReplaceAll(res, "{{PrpNum}}", data.PrpNum)
	res = strings.ReplaceAll(res, "{{PrpDtBeg}}", data.PrpDtBeg)
	res = strings.ReplaceAll(res, "{{PrpDtEnd}}", data.PrpDtEnd)
	res = strings.ReplaceAll(res, "{{Worker}}", data.Worker)
	res = strings.ReplaceAll(res, "{{Row.Name}}", row.Name)
	res = strings.ReplaceAll(res, "{{Row.PayDt}}", row.PayDt)
	res = strings.ReplaceAll(res, "{{Row.Reason}}", row.Reason)
	cell.SetString(res)
}

func xlsRenderRow(row *xlsx.Row, data XlsData, dataRow XlsRow) {
	for _, cell := range row.Cells {
		xlsRenderCell(cell, data, dataRow)
	}
}

func xlsCloneCell(from, to *xlsx.Cell) {
	to.Value = from.Value
	style := from.GetStyle()
	style.ApplyAlignment = true
	if style.Border.Bottom == "" {
		style.Border.Bottom = "none"
	}
	if style.Border.Left == "" {
		style.Border.Left = "none"
	}
	if style.Border.Top == "" {
		style.Border.Top = "none"
	}
	if style.Border.Right == "" {
		style.Border.Right = "none"
	}
	to.SetStyle(style)
	to.HMerge = from.HMerge
	to.VMerge = from.VMerge
	to.Hidden = from.Hidden
	to.NumFmt = from.NumFmt
}

func xlsCloneRow(from, to *xlsx.Row) {
	if from.Height != 0 {
		to.SetHeight(from.Height)
	}

	for _, cell := range from.Cells {
		newCell := to.AddCell()
		xlsCloneCell(cell, newCell)
	}
}

func xlsCloneSheet(from, to *xlsx.Sheet) {
	for _, col := range from.Cols {
		newCol := xlsx.Col{}
		newCol.SetStyle(col.GetStyle())
		newCol.Min = col.Min
		newCol.Max = col.Max
		newCol.Width = col.Width
		newCol.Hidden = col.Hidden
		newCol.Collapsed = col.Collapsed
		to.Cols = append(to.Cols, &newCol)
	}
}

func xlsRecheckWidth(from, to *xlsx.Sheet) {
	for i, col := range from.Cols {
		to.SetColWidth(i, i, col.Width)
	}
}

func XlsRenderTemplate(tfile string, data XlsData, dataRows []XlsRow) (*xlsx.File, error) {
	var source, result *xlsx.File
	var err error
	var sSheet, rSheet *xlsx.Sheet
	var dummyRow XlsRow
	source, err = xlsx.OpenFile(tfile)
	if err != nil {
		return nil, err
	}
	if len(source.Sheets) == 0 {
		return nil, fmt.Errorf("Шаблон %s не содержит листов", tfile)
	}
	sSheet = source.Sheets[0]
	result = xlsx.NewFile()
	rSheet, err = result.AddSheet("Визит")
	if err != nil {
		return nil, err
	}
	xlsCloneSheet(sSheet, rSheet)
	for _, srow := range sSheet.Rows {
		if isIterableRow(srow) {
			for _, dR := range dataRows {
				newrow := rSheet.AddRow()
				xlsCloneRow(srow, newrow)
				xlsRenderRow(newrow, data, dR)
			}
		} else {
			newrow := rSheet.AddRow()
			xlsCloneRow(srow, newrow)
			xlsRenderRow(newrow, data, dummyRow)
		}
	}
	xlsRecheckWidth(sSheet, rSheet)
	return result, nil
}
