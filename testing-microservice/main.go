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
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	logger "github.com/ecpartan/soap-server-tr069/log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ecpartan/soap-server-tr069/internal/config"
	"github.com/ecpartan/soap-server-tr069/internal/parsemap"
)

type RetScriptTask struct {
	Code    string         `json:"code"`
	Message map[string]any `json:"message"`
}

type test_context struct {
	map_inst map[string]string
}

var test_ctx test_context

func get_resp_tsk(i int, mp map[string]any) (map[string]any, error) {
	if analize_task, ok := mp[strconv.Itoa(i+1)]; !ok {
		return nil, errors.New("Not found task1")
	} else {
		if task, ok := analize_task.(map[string]any); !ok {
			return nil, errors.New("Not found task2")
		} else {
			return task, nil
		}
	}
}

func check_code_task(task map[string]any, code string) error {
	tsk_code, ok := task["Code"]
	if !ok {
		return errors.New("Not found code")
	}
	if req_code, ok := tsk_code.(float64); ok {
		resp_port, err := strconv.Atoi(code)
		if err != nil {
			return err
		}
		if resp_port != int(req_code) {
			return errors.New("Code mismatch")
		}
	}
	return nil
}

func substrInst(message string, start, end byte) (bool, int, int) {

	if idx := strings.IndexByte(message, start); idx >= 0 {
		if idx_end := strings.IndexByte(message[idx:], end); idx_end >= 0 {
			return true, idx, idx + idx_end
		} else {
			return true, idx, idx + (idx - len(message) + 1)
		}
	}

	return false, -1, -1
}

func SubstrByToken(str string, token byte, replace_map map[string]string) string {
	if ok, start, end := substrInst(str, token, '.'); ok {
		replacing := str[start:end]
		logger.LogDebug("idx", replacing)

		if replace_trim, ok := replace_map[replacing]; ok {
			return str[:start] + replace_trim + str[end:]
		}
	}

	return ""
}

func check_message(task map[string]any, message map[string]any) error {
	logger.LogDebug("Enter ")

	if find_instance_key, ok := task["Instance"].(string); ok {
		inst := parsemap.GetXMLString(message, "InstanceNumber")
		logger.LogDebug("InstanceNumber22 ", reflect.TypeOf(inst), inst)

		if inst != "" {
			test_ctx.map_inst[find_instance_key] = inst
			return nil
		} else {
			return errors.New("Not found instance")
		}
	}

	if find_instance_path, ok := task["FindInstance"].(string); ok {
		logger.LogDebug("FindInstance22 ", test_ctx.map_inst)

		find_instance_path = SubstrByToken(find_instance_path, '#', test_ctx.map_inst)
		logger.LogDebug("Find instance", find_instance_path)

		paramlist := parsemap.GetXML(message, "ParameterList.ParameterValueStruct").([]any)
		find := false

		for _, v := range paramlist {
			name := parsemap.GetXMLString(v, "Name")
			if name != "" {
				find = strings.HasPrefix(name, find_instance_path)
				if find {
					logger.LogDebug("Finded ", name, find_instance_path)
					return nil
				}
			}
		}
		if !find {
			return fmt.Errorf("Not found name %s in return ParameterValueStruct", find_instance_path)
		}
	}

	if find_inst_values, ok := task["FindValue"].(map[string]any); ok {
		paramlist := parsemap.GetXML(message, "ParameterList.ParameterValueStruct").([]any)

		for path, val := range find_inst_values {
			value_mess := val.(string)
			path = SubstrByToken(path, '#', test_ctx.map_inst)
			find := false
			logger.LogDebug("Find value", path, value_mess)
			for _, v := range paramlist {
				name := parsemap.GetXMLString(v, "Name")
				if name != "" && name == path {
					value_trans := parsemap.GetXMLString(v, "Value")
					if value_trans != "" && value_trans == value_mess {
						logger.LogDebug("Finded ", path, value_mess)
						find = true
					}
				}
			}
			if !find {
				return fmt.Errorf("Not found name %s with value %s in return ParameterValueStruct", path, value_mess)
			}
		}
	}

	return nil
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

	for i, rettask := range arr {
		task, err := get_resp_tsk(i, mp)
		if err != nil {
			return err
		}
		logger.LogDebug("Task: ", task)

		if err := check_code_task(task, rettask.Code); err != nil {
			return err
		}

		err = check_message(task, rettask.Message)

		if err != nil {
			return err
		}

	}

	return nil
}
func mapToString(m map[string]string) string {
	parts := make([]string, 0, len(m))
	parts = append(parts, "Report "+time.Now().Format("2006-01-02 15:04:05")+":	")
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, " ")
}

func SendReport(report map[string]string) {
	bot, err := tgbotapi.NewBotAPI("")
	if err != nil {
		logger.LogDebug("Not found tg bot", err)
		return
	}

	users := []int64{}
	msgText := mapToString(report)

	for _, userID := range users {
		msg := tgbotapi.NewMessage(userID, msgText)
		_, err := bot.Send(msg)
		if err != nil {
			logger.LogDebug("Not sended to tg bot", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	logger.LogDebug("Report sended successfully!")
}

func main() {
	dir := "./scripts"
	files, err := os.ReadDir("./scripts")

	test_ctx = test_context{map_inst: make(map[string]string)}
	report := make(map[string]string)

	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	cfg := config.GetConfig()
	url := fmt.Sprintf("http://%s:%d/integral", cfg.Server.Host, cfg.Server.Port) //"192.168.0.129", 8080)

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

		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		var response []RetScriptTask
		if json.Unmarshal(body, &response) != nil {
			log.Fatalf("JSON err: %v", err)
		}

		err = AnalizeResponse(response, file.Name())
		file_key := strings.Split(file.Name(), ".")[0]
		if err != nil {
			report[file_key] = "FAILED!!! " + err.Error()
		} else {
			report[file_key] = "SUCCESS"
		}

		fmt.Println(err)
	}

	SendReport(report)
}
