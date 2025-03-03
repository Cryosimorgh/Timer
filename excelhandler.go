package main

import (
	"fmt"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

func initExcelFile() {
	if _, err := os.Stat(excelFileName); os.IsNotExist(err) {
		f := excelize.NewFile()
		f.SetCellValue("Sheet1", "A1", "Timestamp")
		f.SetCellValue("Sheet1", "B1", "Event")
		f.SetCellValue("Sheet1", "C1", "Activity Name")
		f.SetCellValue("Sheet1", "D1", "Duration")
		f.SaveAs(excelFileName)
	}
}

func saveToExcel(entry TimerEntry) {
	f, err := excelize.OpenFile(excelFileName)
	if err != nil {
		fmt.Println("Error opening Excel file:", err)
		return
	}
	defer f.Close()

	date := entry.Timestamp.Format("2006-01-02")
	sheetName := date

	// Create a new sheet for a new day if it doesn't exist
	if _, err := f.GetSheetIndex(sheetName); err != nil {
		f.NewSheet(sheetName)
		f.SetCellValue(sheetName, "A1", "Timestamp")
		f.SetCellValue(sheetName, "B1", "Event")
		f.SetCellValue(sheetName, "C1", "Activity Name")
		f.SetCellValue(sheetName, "D1", "Duration")
	}

	rows, _ := f.GetRows(sheetName)
	rowIndex := len(rows) + 1

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), entry.Timestamp)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), entry.Event)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), entry.Name)

	if entry.Duration > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex),
			fmt.Sprintf("%.1f minutes", entry.Duration.Minutes()))
	}

	if err := f.Save(); err != nil {
		fmt.Println("Error: The file is already open.")
		fmt.Println("1. Save to a new file")
		fmt.Println("2. Retry")
		fmt.Println("3. Exit")

		var choice int
		fmt.Print("Enter your choice: ")
		fmt.Scan(&choice)

		switch choice {
		case 1:
			uniqueFilename := getUniqueFilename()
			if err := f.SaveAs(uniqueFilename); err != nil {
				fmt.Println("Error saving Excel file:", err)
			} else {
				fmt.Println("Data saved to:", uniqueFilename)
			}
		case 2:
			if err := saveWithRetry(f, excelFileName, 3); err != nil {
				fmt.Println("Error saving Excel file:", err)
			}
		case 3:
			fmt.Println("Exiting...")
		default:
			fmt.Println("Invalid choice")
		}
	}
}
func getUniqueFilename() string {
	return fmt.Sprintf("report_%s.xlsx", time.Now().Format("20060102_150405"))
}

func saveWithRetry(f *excelize.File, filename string, maxRetries int) error {
	for range maxRetries {
		if err := f.SaveAs(filename); err == nil {
			return nil // Success
		}
		time.Sleep(1 * time.Second) // Wait before retrying
	}
	return fmt.Errorf("failed to save file after %d retries", maxRetries)
}
