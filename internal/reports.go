package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GenerateInventoryReport generates a CSV report from inventory files
func GenerateInventoryReport(inventoryDir string) error {
	files, err := filepath.Glob(filepath.Join(inventoryDir, "*_inventory.txt"))
	if err != nil {
		return fmt.Errorf("error finding inventory files: %w", err)
	}
	
	csvFile, err := os.Create("inventory_report.csv")
	if err != nil {
		return fmt.Errorf("error creating CSV file: %w", err)
	}
	defer csvFile.Close()
	
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()
	
	headers := []string{"Hostname", "Chassis_SN", "Chassis_Age", "RP1_SN", "RP1_Age", "RP2_SN", "RP2_Age", "ESP1_SN", "ESP1_Age", "ESP2_SN", "ESP2_Age", "Avg_Major_Age", "Avg_All_Age"}
	writer.Write(headers)
	
	for _, file := range files {
		basename := filepath.Base(file)
		hostname := strings.TrimSuffix(basename, "_inventory.txt")
		
		chassis, rp, esp, allSerials := ParseInventoryFile(file)
		
		row := []string{hostname}
		
		if len(chassis) > 0 {
			row = append(row, chassis[0], strconv.Itoa(CalculateAge(chassis[0])))
		} else {
			row = append(row, "", "")
		}
		
		for i := 0; i < 2; i++ {
			if i < len(rp) {
				row = append(row, rp[i], strconv.Itoa(CalculateAge(rp[i])))
			} else {
				row = append(row, "", "")
			}
		}
		
		for i := 0; i < 2; i++ {
			if i < len(esp) {
				row = append(row, esp[i], strconv.Itoa(CalculateAge(esp[i])))
			} else {
				row = append(row, "", "")
			}
		}
		
		majorSerials := make([]string, 0)
		majorSerials = append(majorSerials, chassis...)
		majorSerials = append(majorSerials, rp...)
		majorSerials = append(majorSerials, esp...)
		
		avgMajorAge := CalculateAverageAge(majorSerials)
		avgAllAge := CalculateAverageAge(allSerials)
		
		row = append(row, fmt.Sprintf("%.1f", avgMajorAge), fmt.Sprintf("%.1f", avgAllAge))
		
		writer.Write(row)
	}
	
	return nil
}