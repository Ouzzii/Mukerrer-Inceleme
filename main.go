package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

var projectPath, _ = os.Getwd()
var fileName string = "unknown"
var file *excelize.File
var rows [][]string
var err error
var z, n, y int

const (
	Version = "1.1.0"
	User    = "Ouzzii"
	Repo    = "Mukerrer-Inceleme"
)

func init() {
	target := time.Date(2026, time.April, 10, 0, 0, 0, 0, time.Local)
	if time.Now().After(target) {
		os.Exit(1)
	}

	fmt.Printf("Version: %s\n", Version)
	if isUploadAvailable() {
		fmt.Println("Yeni güncelleme bulundu, güncelleme indiriliyor.")
		doUpdate(fmt.Sprintf("https://github.com/%s/%s/releases/latest/download/mukerrer.exe", User, Repo))
		fmt.Println("Güncelleme tamamlandı.")
	}

	err := os.Mkdir("mükerrer", os.ModePerm)
	if err != nil {
		if !strings.Contains(err.Error(), "Cannot create a file when that file already exists.") {
			Closefunc(err)
		}
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Incident listesini giriniz: ")
	fileName, _ = reader.ReadString('\n')

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
}

func main() {

	rows, err = file.GetRows("Incidents listesi")
	if err != nil {
		Closefunc(err)
		return
	}

	for n, frow := range rows[1:] {
		for y, lrow := range rows[1:] {
			if frow[3] == lrow[3] && frow[18] < lrow[19] && frow[19] > lrow[18] && frow[22] == "ALL" && lrow[22] != "ALL" {
				fmt.Println(frow[0], lrow[0])
				cell, _ := excelize.CoordinatesToCellName(1, n+2)
				newrow := append(frow, "mükerrer_all")
				file.SetSheetRow("Incidents listesi", cell, &newrow)

				cell, _ = excelize.CoordinatesToCellName(1, y+2)
				newrow = append(lrow, "mükerrer")
				file.SetSheetRow("Incidents listesi", cell, &newrow)
				fmt.Printf("Mükerrer Bulundu: %s - %s: %s\n", frow[0], lrow[0], frow[3])

			} else if frow[3] == lrow[3] && frow[18] < lrow[19] && frow[19] > lrow[18] && frow[22] == lrow[22] && frow[0] != lrow[0] {
				fmt.Println(frow[0], lrow[0], frow[22], lrow[22])
				cell, _ := excelize.CoordinatesToCellName(1, n+2)
				newrow := append(frow, "Mükerrer1")
				file.SetSheetRow("Incidents listesi", cell, &newrow)

				cell, _ = excelize.CoordinatesToCellName(1, y+2)
				newrow = append(lrow, "Mükerrer2")
				file.SetSheetRow("Incidents listesi", cell, &newrow)
				fmt.Printf("Mükerrer Bulundu: %s - %s: %s\n", frow[0], lrow[0], frow[3])
			}

		}

	}

	var newfilename = filepath.Join(projectPath, "mükerrer", strings.Replace(filepath.Base(fileName), ".", fmt.Sprintf(" %s.", time.Now().Format("02-01-2006")), -1))
	file.SaveAs(newfilename)
	fmt.Printf("Yeni dosya %s adı ile kaydedilmiştir.", newfilename)

	os.Exit(1)

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
	/*if _, err := os.Create(old_path); err != nil {
		log.Fatalf("failed to create dummy file: %v", err)
	}*/
	//fmt.Printf("Created dummy file: %s\n", old_path)

	// Get the directory path
	dir := filepath.Dir(old_path)

	// Get the file name without the extension
	filenameWithoutExt := strings.TrimSuffix(filepath.Base(old_path), filepath.Ext(old_path))

	// Construct the new file name
	newFileName := filenameWithoutExt + "." + new_ext

	// Construct the full new path
	newPath := filepath.Join(dir, newFileName)

	// Rename the file
	//err := os.Rename(old_path, newPath)
	if err != nil {
		Closefunc(err)
	}
	return newPath

}

func Closefunc(err error) {
	if err != nil {
		fmt.Printf("Hata nedeniyle program çalışması durdurulmuştur: %s\n", err.Error())

		// Dosya ve satır bilgisini göster
		_, file, line, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("Hata burada oluştu: %s:%d\n", file, line)
		}

		// Log klasörünü oluştur (eğer yoksa)
		os.MkdirAll("log", os.ModePerm)

		// Log dosyası ismi: log/dosya_adi.txt
		logPath := filepath.Join(projectPath, "log", filepath.Base(fileName+".txt"))
		f, fErr := os.Create(logPath)
		if fErr != nil {
			fmt.Println("Log dosyası oluşturulamadı:", fErr)
		} else {
			defer f.Close()
			// Sadece stack trace kaydet
			f.Write(debug.Stack())
			fmt.Println("Stack trace loglandı:", logPath)
		}

		// Console stack trace
		fmt.Println("Stack trace:")
		debug.PrintStack()
	}

	// Geri sayım
	for i := 10; i > 0; i-- {
		fmt.Printf("Program %d saniye sonra kapanacaktır\n", i)
		time.Sleep(1 * time.Second)
	}
	os.Exit(1)
}

