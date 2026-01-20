package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type RetScriptTask struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func AnalizeResponse(arr []RetScriptTask, path string) error {
	dir := "./analize"
	filePath := filepath.Join(dir, path)

	content, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer content.Close()
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	mp := make(map[string]any)
	if json.Unmarshal(data, &mp) != nil {
		return err
	}

	//map_instance:= make(map[string]string)
	for i, rettask := range arr {
		if analize_task, ok := mp[strconv.Itoa(i+1)]; !ok {
			return errors.New("Not found task")
		} else {
			if task, ok := analize_task.(map[string]any); !ok {
				return errors.New("Not found task")
			} else {
				if task["code"] != rettask.Code {
					return errors.New("Not found task")
				}

			}

		}
	}

	fmt.Println(mp)

	return nil
}

func main() {
	dir := "./scripts"
	files, err := os.ReadDir("./scripts")

	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	//cfg := config.GetConfig()
	url := fmt.Sprintf("http://%s:%d/integral", "192.168.0.129", 8080)

	client := &http.Client{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, file.Name())

		content, err := os.Open(filePath)
		if err != nil {
			log.Printf("Error open file %s: %v\n", filePath, err)
			continue
		}
		defer content.Close()

		data, err := io.ReadAll(content)
		if err != nil {
			log.Printf("Ошибка при чтении файла %s: %v\n", filePath, err)
			continue
		}

		mp := make(map[string]any)
		if json.Unmarshal(data, &mp) != nil {
			log.Fatalf("JSON err: %v", err)
		}

		jsonData, err := json.Marshal(mp)
		if err != nil {
			log.Fatalf("JSON err: %v", err)
		}

		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Fatal(err)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		response := make([]RetScriptTask, 0)
		if json.Unmarshal(body, &response) != nil {
			log.Fatalf("JSON err: %v", err)
		}

		AnalizeResponse(response, file.Name())

		fmt.Println(string(body))

	}

}
