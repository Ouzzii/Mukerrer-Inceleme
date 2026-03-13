package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/schollz/progressbar/v3"
	"github.com/xuri/excelize/v2"
)

//go:embed template/template.xlsx
var embedTemplateFile []byte

var projectPath, _ = os.Getwd()
var fileName string = "unknown"
var file, templateFile *excelize.File
var rows [][]string
var err error
var n, y int
var layout = "2.1.2006 15:04:05"

const (
	Version = "1.2.0"
	User    = "Ouzzii"
	Repo    = "Mukerrer-Inceleme"
)

var args struct {
	Filename string `arg:"positional"`
	Log      bool
}

func init() {
	arg.MustParse(&args)
	target := time.Date(2026, time.April, 10, 0, 0, 0, 0, time.Local)
	if time.Now().After(target) {
		os.Exit(1)
	}
	cleanupOldExe()
	fmt.Printf("Version: %s\n", Version)
	if isUploadAvailable() {
		fmt.Println("Yeni güncelleme bulundu, güncelleme indiriliyor.")
		if err := doUpdate(fmt.Sprintf("https://github.com/%s/%s/releases/latest/download/mukerrer.exe", User, Repo)); strings.Contains(err.Error(), "dial tcp: lookup github.com: no such host") {
			fmt.Println("İnternet bağlantısı bulunmadığından güncelleme indirilemedi.")
			Closefunc(err)
		} else if err != nil {
			Closefunc(err)
		}
		fmt.Println("Güncelleme tamamlandı.")
	}

	err := os.Mkdir("mükerrer", os.ModePerm)
	if err != nil {
		if !strings.Contains(err.Error(), "Cannot create a file when that file already exists.") {
			Closefunc(err)
		}
	}
	if len(args.Filename) < 5 {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Incident listesini giriniz: ")
		fileName, _ = reader.ReadString('\n')
	} else {
		fileName = args.Filename
		fmt.Printf("%s isimili dosya seçildi\n", fileName)
	}

	fileName = strings.TrimSpace(fileName)
	fileName = strings.TrimPrefix(fileName, "& ")
	fileName = strings.Trim(fileName, "\"'")

	excelType := detectExcelType(fileName)
	switch excelType {
	case "xls":
		fmt.Println("Dosya formatı xlsx biçiminde değildir, xls formatından xlsx biçimine çeviriliyor.")
		if err := convertXLS(fileName, change_file_extension(fileName, "xlsx")); err != nil {
			Closefunc(err)
		} else {
			fmt.Println("Dosya biçim dönüştürme işlemi tamamlandı: xls -> xlsx")
			fileName = change_file_extension(fileName, "xlsx")
		}
	case "html":
		fmt.Println("Dosya formatı xlsx biçiminde değildir, html formatından xlsx biçimine çeviriliyor.")
		if err := convertHTML(fileName, change_file_extension(fileName, "xlsx")); err != nil {
			Closefunc(err)
		} else {
			fmt.Println("Dosya biçim dönüştürme işlemi tamamlandı: html -> xlsx")
			fileName = change_file_extension(fileName, "xlsx")
		}
	case "xlsx":
		fmt.Println("xlsx biçiminde dosya bulundu.")
	}

	file, err = excelize.OpenFile(fileName)

	if err != nil {
		Closefunc(err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			Closefunc(err)
		}
	}()

	//templateFile, err = excelize.OpenFile("./template/template.xlsx")
	templateFile, err = excelize.OpenReader(bytes.NewReader(embedTemplateFile))

	if err != nil {
		Closefunc(err)
		return
	}

}

func main() {

	rows, err = file.GetRows("Incidents listesi")
	if err != nil {
		Closefunc(err)
		return
	}
	header := append(rows[0], "Saha Cell", "KESİNTİ BÇO", "Cell Sayısı", "Toplam Kesinti Süresi Dakika", "Toplam Kesinti Süresi Cell Saniye", "Mükerrer", "Toplam Kesinti Saha Cell")
	templateFile.SetSheetRow("Incidents listesi", "A1", &header)
	bar := progressbar.Default(int64(len(rows[1:]) * len(rows[1:])))

	for n, frow := range rows[1:] {

		for y, lrow := range rows[1:] {

			if frow[3] == lrow[3] && frow[18] < lrow[19] && frow[19] > lrow[18] && frow[22] == "ALL" && lrow[22] != "ALL" {

				templateFile.SetCellValue("Incidents listesi", fmt.Sprintf("AL%d", n+2), "mükerrer_all")

				templateFile.SetCellValue("Incidents listesi", fmt.Sprintf("AL%d", y+2), "mükerrer")

			} else if frow[3] == lrow[3] && frow[18] < lrow[19] && frow[19] > lrow[18] && frow[22] == lrow[22] && frow[0] != lrow[0] {

				templateFile.SetCellValue("Incidents listesi", fmt.Sprintf("AL%d", n+2), "mükerrer1")

				templateFile.SetCellValue("Incidents listesi", fmt.Sprintf("AL%d", y+2), "mükerrer2")
			}
			bar.Add(1)

		}

		templateFile.SetSheetRow("Incidents listesi", fmt.Sprintf("A%d", n+2), &frow)

		R, err := time.Parse(layout, cleanDate(frow[17]))
		if err != nil {
			Closefunc(err)
		}
		templateFile.SetCellValue("Incidents listesi", fmt.Sprintf("R%d", n+2), R)
		S, err := time.Parse(layout, cleanDate(frow[18]))
		if err != nil {
			fmt.Println([]byte(frow[18]))
			Closefunc(err)
		}
		templateFile.SetCellValue("Incidents listesi", fmt.Sprintf("S%d", n+2), S)
		T, err := time.Parse(layout, cleanDate(frow[19]))
		if err != nil {
			Closefunc(err)
		}
		templateFile.SetCellValue("Incidents listesi", fmt.Sprintf("T%d", n+2), T)
		U, err := time.Parse(layout, cleanDate(frow[20]))
		if err != nil {
			Closefunc(err)
		}
		templateFile.SetCellValue("Incidents listesi", fmt.Sprintf("U%d", n+2), U)
		templateFile.SetCellFormula("Incidents listesi", fmt.Sprintf("AG%d", n+2), fmt.Sprintf(`=IF(W%d="ALL",COUNTIFS(Cell!B:B,D%d,Cell!C:C,"ACTIVE"),1)`, n+2, n+2))

		templateFile.SetCellFormula("Incidents listesi", fmt.Sprintf("AH%d", n+2), fmt.Sprintf(`=VLOOKUP(N%d,'Kesinti Sorumluluğu'!A:B,2,0)`, n+2))
		templateFile.SetCellFormula("Incidents listesi", fmt.Sprintf("AI%d", n+2), fmt.Sprintf(`=IF(W%d="ALL",COUNTIFS(Cell!B:B,D%d,Cell!F:F,"ACTIVE"),1)`, n+2, n+2))
		templateFile.SetCellFormula("Incidents listesi", fmt.Sprintf("AJ%d", n+2), fmt.Sprintf(`=IF(T%d="",0,T%d-S%d)`, n+2, n+2, n+2))
		templateFile.SetCellFormula("Incidents listesi", fmt.Sprintf("AK%d", n+2), fmt.Sprintf(`=AI%d*AJ%d`, n+2, n+2))

		templateFile.SetCellFormula("Incidents listesi", fmt.Sprintf("AM%d", n+2), fmt.Sprintf(`=AG%d*AJ%d`, n+2, n+2))

	}

	format := "[m]"

	style, err := templateFile.NewStyle(&excelize.Style{
		CustomNumFmt: &format,
	})
	if err != nil {
		Closefunc(err)
	}

	err = templateFile.SetColStyle("Incidents listesi", "AJ", style)
	if err != nil {
		Closefunc(err)
	}

	err = templateFile.SetColStyle("Incidents listesi", "AK", style)
	if err != nil {
		Closefunc(err)
	}

	format = "[s]"

	style, err = templateFile.NewStyle(&excelize.Style{
		CustomNumFmt: &format,
	})
	if err != nil {
		Closefunc(err)
	}
	err = templateFile.SetColStyle("Incidents listesi", "AM", style)
	if err != nil {
		Closefunc(err)
	}

	if err := templateFile.UpdateLinkedValue(); err != nil {
		log.Println("UpdateLinkedValue hatası:", err)
	}
	var newfilename = filepath.Join(projectPath, "mükerrer", strings.Replace(filepath.Base(fileName), ".", fmt.Sprintf(" %s.", time.Now().Format("02-01-2006")), -1))

	AutoFitColumns(templateFile, "Incidents listesi")
	templateFile.SaveAs(newfilename)
	fmt.Printf("Yeni dosya %s adı ile kaydedilmiştir.\n", newfilename)
	Closefunc(nil)

}

func detectExcelType(path string) string {

	f, err := os.Open(path)
	if err != nil {
		return "unknown"
	}
	defer f.Close()

	buf := make([]byte, 4096)
	n, err := f.Read(buf)
	if err != nil {
		return "unknown"
	}

	data := buf[:n]

	// XLSX / XLSM / XLSB (zip container)
	if bytes.HasPrefix(data, []byte{0x50, 0x4B, 0x03, 0x04}) {
		return "xlsx"
	}

	// Old Excel XLS (OLE compound)
	if bytes.HasPrefix(data, []byte{0xD0, 0xCF, 0x11, 0xE0}) {
		return "xls"
	}

	// HTML Excel
	if bytes.Contains(bytes.ToLower(data), []byte("<html")) ||
		bytes.Contains(bytes.ToLower(data), []byte("<table")) {
		return "html"
	}

	// Excel XML format
	if bytes.Contains(data, []byte("<Workbook")) {
		return "xml"
	}

	// CSV tahmini
	if bytes.Contains(data, []byte(",")) && bytes.Contains(data, []byte("\n")) {
		return "csv"
	}

	return "unknown"
}

func change_file_extension(old_path, new_ext string) string {

	dir := filepath.Dir(old_path)

	filenameWithoutExt := strings.TrimSuffix(filepath.Base(old_path), filepath.Ext(old_path))

	newFileName := filenameWithoutExt + "." + new_ext

	newPath := filepath.Join(dir, newFileName)

	if err != nil {
		Closefunc(err)
	}
	return newPath

}

func Closefunc(err error) {
	if err != nil && args.Log == true {
		fmt.Printf("Hata nedeniyle program çalışması durdurulmuştur: %s\n", err.Error())

		_, file, line, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("Hata burada oluştu: %s:%d\n", file, line)
		}

		os.MkdirAll("log", os.ModePerm)

		logPath := filepath.Join(projectPath, "log", filepath.Base(fileName+".txt"))
		f, fErr := os.Create(logPath)
		if fErr != nil {
			fmt.Println("Log dosyası oluşturulamadı:", fErr)
		} else {
			defer f.Close()
			f.Write(debug.Stack())
			fmt.Println("Stack trace loglandı:", logPath)
		}

		fmt.Println("Stack trace:")
		debug.PrintStack()
	}

	cleanupOldExe()

	if templateFile != nil {
		if err := templateFile.Close(); err != nil {
			fmt.Println("Dosya kapatılırken hata:", err)
		}
	}

	for i := 10; i > 0; i-- {
		fmt.Printf("Program %d saniye sonra kapanacaktır\n", i)
		time.Sleep(1 * time.Second)
	}
	os.Exit(1)
}

func cleanDate(s string) string {
	s = strings.ReplaceAll(s, "\u00A0", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.TrimSpace(s)
	return s
}

func AutoFitColumns(f *excelize.File, sheet string) error {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return err
	}

	colWidths := map[int]float64{}

	for _, row := range rows {
		for colIndex, cell := range row {

			length := float64(len(cell))*1.1 + 2

			if length > colWidths[colIndex] {
				colWidths[colIndex] = length
			}
		}
	}

	for colIndex, width := range colWidths {

		colName, err := excelize.ColumnNumberToName(colIndex + 1)
		if err != nil {
			return err
		}

		err = f.SetColWidth(sheet, colName, colName, width+2)
		if err != nil {
			return err
		}
	}

	return nil
}
