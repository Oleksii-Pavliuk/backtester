package processor

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Processor struct {
	outputDir string
}

func NewProcessor(outputDir string, host string, port int) *Processor {
	p := Processor{outputDir: outputDir}
	if _, err := os.ReadDir(p.outputDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(p.outputDir, 0755); err != nil {
				log.Fatalf("Failed to create output directory %s: %v", p.outputDir, err)
			}
			log.Printf("Created directory: %s", p.outputDir)
		} else {
			log.Fatalf("Failed to read directory %s: %v", p.outputDir, err)
		}
	}

	api := Api{
		Address:   host,
		Port:      port,
		Processor: &p,
		handlers:  make(map[string]time.Time)}
	api.Start()
	return &p
}

func (p *Processor) Convert(inputFile []byte) (string, error) {

	data, err := p.readCSV(inputFile)
	if err != nil {
		return "", fmt.Errorf("Error reading file", err)
	}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	filename := uuid.New().String()

	outputFile := path.Join(p.outputDir, filename+".json")

	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write JSON file: %v", err)
	}

	fmt.Printf("Successfully converted CSV to JSON: %s\n", outputFile)
	return filename, nil
}

func (p *Processor) readCSV(inputFile []byte) ([]map[string]any, error) {

	reader := csv.NewReader(bytes.NewReader(inputFile))

	headersRequired := []string{"timestamp", "open", "high", "low", "close", "volume"}

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}
	if reflect.DeepEqual(headers, headersRequired) {
		fmt.Println("Headers match!")
	} else {
		return nil, fmt.Errorf("Headers do not match!")
	}

	var data []map[string]any

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %v", err)
		}

		record := make(map[string]any)
		for i, value := range row {
			if headers[i] == "timestamp" {
				record[headers[i]], err = p.convertToDate(value)
			} else {
				record[headers[i]], err = p.convertToFloat(value)
			}
		}
		data = append(data, record)
	}

	data = p.sortByTimestamp(data)
	return data, nil
}

func (p *Processor) readJSON(id string) ([]map[string]interface{}, error) {
	filePath := path.Join(p.outputDir, id+".json")

	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File does not exist: %s", filePath)
			return nil, fmt.Errorf("file not found: %s", filePath)
		}
		log.Fatalf("Failed to read file %s: %v", filePath, err)
		return nil, err
	}

	var data []map[string]interface{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		log.Printf("Failed to parse JSON: %v", err)
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return data, nil
}

func (p *Processor) sortByTimestamp(data []map[string]any) []map[string]any {

	sort.SliceStable(data, func(i, j int) bool {
		timestamp1, okI := data[i]["timestamp"].(time.Time)
		timestamp2, okJ := data[j]["timestamp"].(time.Time)

		if !okI || !okJ {
			return false
		}
		return timestamp1.Before(timestamp2)
	})
	return data
}

func (p *Processor) convertToDate(value string) (time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04:05", value)
	if err != nil {
		return t, fmt.Errorf("invalid timestamp: %v", err)
	}
	return t, nil
}

func (p *Processor) convertToFloat(value string) (float64, error) {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float: %v", err)
	}
	return floatValue, nil
}

func (p *Processor) findRecordByTimestamp(data []map[string]interface{}, timestamp string) (map[string]interface{}, int, bool) {
	for index, row := range data {
		if ts, ok := row["timestamp"].(string); ok && ts == timestamp {
			return row, index, true
		}
	}
	return nil, 0, false
}
