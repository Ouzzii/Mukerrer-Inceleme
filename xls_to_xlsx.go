package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/extrame/xls"
	"github.com/xuri/excelize/v2"
)

func convertXLS(input, output string) error {

	xlsFile, err := xls.Open(input, "utf-8")
	if err != nil {
		Closefunc(err)
		return err
	}

	f := excelize.NewFile()

	for i := 0; i < xlsFile.NumSheets(); i++ {

		sheet := xlsFile.GetSheet(i)
		if sheet == nil {
			continue
		}

		sheetName := sheet.Name

		if i == 0 {
			f.SetSheetName("Sheet1", sheetName)
		} else {
			f.NewSheet(sheetName)
		}

		for r := 0; r <= int(sheet.MaxRow); r++ {

			row := sheet.Row(r)
			if row == nil {
				continue
			}

			colCount := row.LastCol()
			rowData := make([]interface{}, colCount)

			for c := 0; c < colCount; c++ {
				rowData[c] = row.Col(c)
			}

			cell, _ := excelize.CoordinatesToCellName(1, r+1)

			f.SetSheetRow(sheetName, cell, &rowData)
		}
	}

	err = f.SaveAs(output)
	if err != nil {
		Closefunc(err)
		return err
	}

	return nil
}

func convertHTML(input, output string) error {
	data, err := os.ReadFile(input)
	if err != nil {
		Closefunc(err)
		return err
	}
	re := regexp.MustCompile(`<x:Name>(.*?)</x:Name>`)
	matches := re.FindAllStringSubmatch(string(data), -1)

	sheetNames := []string{}
	for _, m := range matches {
		if len(m) > 1 {
			sheetNames = append(sheetNames, m[1])
		}
	}
	if len(sheetNames) == 0 {
		sheetNames = append(sheetNames, "Sheet1")
	}

	// HTML’i goquery ile parse et
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	xlsx := excelize.NewFile()
	firstSheet := xlsx.GetSheetName(xlsx.GetActiveSheetIndex())

	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		sheetName := ""
		if i < len(sheetNames) {
			sheetName = sheetNames[i]
		} else {
			sheetName = fmt.Sprintf("Sheet%d", i+1)
		}

		if i == 0 {
			xlsx.SetSheetName(firstSheet, sheetName)
		} else {
			xlsx.NewSheet(sheetName)
		}

		rowIndex := 1
		table.Find("tr").Each(func(_ int, tr *goquery.Selection) {
			colIndex := 1
			tr.Find("td, th").Each(func(_ int, td *goquery.Selection) {
				text := strings.TrimSpace(td.Text())
				cell, _ := excelize.CoordinatesToCellName(colIndex, rowIndex)
				xlsx.SetCellValue(sheetName, cell, text)
				colIndex++
			})
			rowIndex++
		})
	})

	return xlsx.SaveAs(output)
}
