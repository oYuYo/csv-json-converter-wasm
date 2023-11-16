package main

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"syscall/js"
)

type FileType int

const (
	CSV FileType = iota
	JSON
)

// args[0]: FileType, args[1]: fileName, args[2]: textContent
func SelectFunc(this js.Value, args []js.Value) interface{} {
	if len(args) != 3 {
		printAlert("引数の数が不正です")
		return nil
	}

	var tmp string = args[0].String()
	f, err := strconv.Atoi(tmp)
	if err != nil {
		printAlert("引数が不正です")
		return nil
	}
	if f < 0 || 1 < f {
		printAlert("不正なファイル形式です")
		return nil
	}
	fileType := FileType(f)

	fileName := args[1].String()
	ext := filepath.Ext(fileName)
	if (ext == "csv" && fileType != CSV) || (ext == "json" && fileType != JSON) {
		printAlert("不正なファイル形式です")
		return nil
	}

	var text string = args[2].String()
	if fileType == CSV {
		err := CSVtoJSON(&text, strings.TrimSuffix(fileName, ext))
		if err != nil {
			//printAlert(err.Error())
			printAlert("JSONファイルへの変換に失敗しました")
			return nil
		}
	} else {
		err := JSONtoCSV(&text, strings.TrimSuffix(fileName, ext))
		if err != nil {
			printAlert("CSVファイルへの変換に失敗しました")
			return nil
		}
	}
	return nil
}

func CSVtoJSON(text *string, fileName string) error {
	reader := csv.NewReader(strings.NewReader(*text))
	//reader.FieldsPerRecord = -1 //カラム数が可変であることを示す
	header, err := reader.Read()
	if err != nil {
		return err
	}

	var jsonData []map[string]string

	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		data := make(map[string]string)

		for i, value := range row {
			if i < len(header) {
				data[header[i]] = value
			}
		}
		jsonData = append(jsonData, data)
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	attachData(jsonBytes, fileName, ".json")

	return nil
}

func JSONtoCSV(text *string, fileName string) error {
	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(*text), &data); err != nil {
		return err
	}

	var csvBuilder strings.Builder
	csvWriter := csv.NewWriter(&csvBuilder)

	header := make([]string, 0)
	for key := range data[0] {
		header = append(header, key)
	}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	for _, record := range data {
		row := make([]string, 0)
		for _, key := range header {
			value, exists := record[key]
			if exists {
				row = append(row, fmt.Sprintf("%v", value))
			} else {
				row = append(row, "")
			}
		}

		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return err
	}

	attachData([]byte(csvBuilder.String()), fileName, ".csv")

	return nil
}

func attachData(data []byte, fileName string, ext string) {
	document := js.Global().Get("document")
	el := document.Call("getElementById", "output-file")
	encode := base64.StdEncoding.EncodeToString(data)
	dataUri := fmt.Sprintf("data:%s;base64,%s", "text/csv", encode)
	el.Set("href", dataUri)
	el.Set("download", fileName+ext)
	el.Call("click")
}

func printAlert(msg string) {
	document := js.Global().Get("document")
	el := document.Call("getElementById", "err-msg-spn")
	el.Set("innerText", msg)
}

func main() {
	ch := make(chan struct{}, 0)
	js.Global().Set("SelectFunc", js.FuncOf(SelectFunc))
	<-ch
}
