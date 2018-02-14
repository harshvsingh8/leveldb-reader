package main

import (
	"bytes"
	"strings"
	"encoding/json"
	"encoding/hex"
	"os"
	"fmt"
)

import "github.com/syndtr/goleveldb/leveldb"

func splitToJsonStrArray(str string) string {
	var fmtBuffer bytes.Buffer
	lines := strings.Split(str, "\n")
	for i, line := range lines {
		fmtBuffer.WriteString(fmt.Sprintf("\"%s\"", line))
		if i < len(lines) - 1 {
			fmtBuffer.WriteString(",")
		}
	}
	return fmtBuffer.String()
}

func main() {

	printUsage := func() {
		fmt.Println("Usage: leveldb_readkey db_folder_path key1 [key2] [key3]")
	}

	fileExists := func(path string) (bool, error) {
		_, err := os.Stat(path)
		if err == nil { return true, nil }
		if os.IsNotExist(err) { return false, nil }
		return true, err
	}

	if len(os.Args) == 1 {
		fmt.Println("Level/Snappy DB folder path is not supplied")
		printUsage()
		return
	}

	dbPath := os.Args[1]

	dbPresent, err := fileExists(dbPath)

	if !dbPresent {
		fmt.Printf("The DB path: %s does not exist.\n", dbPath)
		printUsage()
		return
	}

	db, err := leveldb.OpenFile(dbPath, nil)
	defer db.Close()

	if err != nil {
		fmt.Println("Could not open DB from:", dbPath)
		printUsage()
		return
	}

	formatValue := func(data []byte) string {
		dataStr := string(data[:])
		var dataMap map[string]interface{}
		err := json.Unmarshal([]byte(dataStr), &dataMap)
		if err != nil {
			// fallback encoding (dumping raw text and then to hex)
			nonJsonData := fmt.Sprintf("{\"__STR\":\"%s\"}", strings.Replace(dataStr,"\n", "\\n", -1))
			// fmt.Println(nonJsonData)
			err = json.Unmarshal([]byte(nonJsonData), &dataMap)
			if err != nil {
				hexDumpStr := hex.Dump(data)
				hexLinesJsonStr := splitToJsonStrArray(hexDumpStr)
				nonJsonData := fmt.Sprintf("{\"__HEX\":[%s]}", hexLinesJsonStr)
				err = json.Unmarshal([]byte(nonJsonData), &dataMap)
				if err != nil {
					return fmt.Sprintf("\"Error Unmarshalling:%s, %s\"", nonJsonData, err)
				}
			}
		}
		jsonData, err := json.Marshal(dataMap)
		if err != nil { return fmt.Sprintf("\"Error Marshalling:%s, %s\"", dataStr, err) }
		var out bytes.Buffer 
		json.Indent(&out, jsonData, "  ", "  ")
		return out.String()
	}

	printKey := func(key string) {
		data, err := db.Get([]byte(key), nil)
		if err != nil {
			fmt.Println("Error reading Key:", key)
			return
		}
		fmt.Printf("  \"%s\":%s",key, formatValue(data))
	}

	fmt.Print("{")
	for i, value := range os.Args[2:] {
		fmt.Println()
		printKey(value)
		if i + 3 != len(os.Args) {
			fmt.Print(",")
		}
	}
	fmt.Println("\n}")
}
