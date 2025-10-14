package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ecpartan/soap-server-tr069/internal/config"
	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"

	"github.com/ecpartan/soap-server-tr069/server"
	"github.com/spf13/cobra"
)

type RunServer struct {
	s *server.Server
}

var rootCmd = &cobra.Command{
	Use:   "soap-server-tr069",
	Short: "A simple SOAP server for TR-069 devices",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cfg := config.GetConfig()
		logger.InitLogger(os.Stdout)

		s, err := server.NewServer(ctx, cfg)
		if err != nil {
			return nil
		}
		ss := RunServer{s: s}

		cmd.SetContext(context.WithValue(cmd.Context(), "server", &ss))

		if err != nil {
			logger.LogDebug("error creating server", err)
			return nil
		}
		s.Register()

		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start TR-069 server",
	Run: func(cmd *cobra.Command, args []string) {
		if ss, ok := cmd.Context().Value("server").(*RunServer); ok {
			logger.LogDebug("ss", ss)
			err := ss.s.Run(context.Background())
			if err != nil {
				logger.LogDebug("error running server", err)
			}
		}
	},
}

func recurse(lst map[string]any, curr string, arr *[]*string) {

	if val, ok := lst["Value"]; ok {
		addObj := curr + ":" + val.(string)
		*arr = append(*arr, &addObj)
	}

	for k, v := range lst {

		if mp, ok := v.(map[string]any); ok {

			if len(k) == 0 {
				continue
			}

			if val, ok := mp["Value"]; ok {
				addObj := curr + k + ":" + val.(string)
				*arr = append(*arr, &addObj)

			} else {
				curr += k + "."

				recurse(mp, curr, arr)
				curr = curr[:len(curr)-len(k)-1]
			}

		}
	}
}

var gettreeCmd = &cobra.Command{
	Use:   "getTree",
	Short: "Execute GetParametervalues RPC for a given device by serial number",
	Args:  cobra.ExactArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		sn := args[0]

		c := repository.NewCache(context.Background(), config.GetConfig())
		tree := c.Get(sn)

		fmt.Println(sn)
		var curr string

		var arr = []*string{}

		recurse(tree, curr, &arr)

		var result string
		for _, line := range arr {
			result += *line + "\n"
		}
		fmt.Println(result)
	},
}

var getvalueCmd = &cobra.Command{
	Use:   "getValue",
	Short: "Execute GetParametervalues cache value for a given device by serial number",
	Args:  cobra.ExactArgs(2),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		sn := args[0]
		path := args[1]

		c := repository.NewCache(context.Background(), config.GetConfig())
		tree := c.Get(sn)
		tree_path := p.GetXML(tree, path)

		var curr string

		var arr = []*string{}

		if mp, ok := tree_path.(map[string]any); ok {
			recurse(mp, curr, &arr)
		}

		var result string
		for _, line := range arr {
			result += path + *line + "\n"
		}
		fmt.Println(result)
	},
}

type getExecStruct struct {
	Script struct {
		Num1 struct {
			GetParameterValues struct {
				Name []string `json:"Name"`
			} `json:"GetParameterValues"`
		} `json:"1"`
		Serial string `json:"Serial"`
	} `json:"Script"`
}

var getvalueExecCmd = &cobra.Command{
	Use:   "getValueExec",
	Short: "Execute GetParametervalues RPC for a given device by serial number",
	Args:  cobra.ExactArgs(2),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		sn := args[0]
		path := args[1]

		script := getExecStruct{}
		script.Script.Num1.GetParameterValues.Name = []string{path}
		script.Script.Serial = sn

		cfg := config.GetConfig()
		url := fmt.Sprintf("http://%s:%d/frontcli", cfg.Server.Host, cfg.Server.Port)

		client := &http.Client{}

		jsonData, err := json.Marshal(script)
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
		fmt.Println(string(body))

		c := repository.NewCache(context.Background(), cfg)
		tree := c.Get(sn)

		var find_path string

		if path[:len(path)-1] == "." {
			find_path = path[:len(path)-1]
		} else {
			find_path = path
		}

		tree_path := p.GetXML(tree, find_path)
		fmt.Println(tree_path)

		var curr string

		var arr = []*string{}

		if mp, ok := tree_path.(map[string]any); ok {
			recurse(mp, curr, &arr)
		}

		var result string
		for _, line := range arr {
			result += path + *line + "\n"
		}
		fmt.Println(result)
	},
}

type setParameterValueStruct struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type setExecStruct struct {
	Script struct {
		Num1 struct {
			SetParameterValueStruct []setParameterValueStruct `json:"SetParameterValues"`
		} `json:"1"`
		Serial string `json:"Serial"`
	} `json:"Script"`
}

var setvalueExecCmd = &cobra.Command{
	Use:   "setValueExec",
	Short: "Execute GetParametervalues RPC for a given device by serial number",
	Args:  cobra.MaximumNArgs(5),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		sn := args[0]

		cfg := config.GetConfig()
		c := repository.NewCache(context.Background(), cfg)
		tree := c.Get(sn)

		script := setExecStruct{}
		script.Script.Serial = sn

		for i, path := range args[1:] {
			fmt.Println(i, path)

			set_args := strings.Split(path, "=")
			name := set_args[0]
			val := set_args[1]
			name_type := p.GetXMLType(tree, name)
			script.Script.Num1.SetParameterValueStruct = append(script.Script.Num1.SetParameterValueStruct, setParameterValueStruct{Name: name, Value: val, Type: name_type})
		}

		url := fmt.Sprintf("http://%s:%d/frontcli", cfg.Server.Host, cfg.Server.Port)

		client := &http.Client{}

		jsonData, err := json.Marshal(script)
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
		fmt.Println(string(body))

		var arr = []*string{}

		for _, setpath := range args[1:] {
			set_args := strings.Split(setpath, "=")
			path := set_args[0]
			tree_path := p.GetXML(tree, path)

			var curr string
			if mp, ok := tree_path.(map[string]any); ok {
				recurse(mp, curr, &arr)
			}
		}
		var result string
		for i, line := range arr {
			set_args := strings.Split(args[i+1], "=")
			path := set_args[0]
			result += path + *line + "\n"
		}
		fmt.Println(result)
	},
}

var execScriptCmd = &cobra.Command{
	Use:   "execScript",
	Short: "Execute GetParametervalues RPC for a given device by serial number",
	Args:  cobra.ExactArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		content, err := os.ReadFile(filename)
		if err != nil {
			return
		}

		mp := make(map[string]any)
		if json.Unmarshal(content, &mp) != nil {
			return
		}

		cfg := config.GetConfig()
		url := fmt.Sprintf("http://%s:%d/frontcli", cfg.Server.Host, cfg.Server.Port)

		client := &http.Client{}

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
		fmt.Println(body)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(gettreeCmd)
	rootCmd.AddCommand(execScriptCmd)
	rootCmd.AddCommand(getvalueCmd)
	rootCmd.AddCommand(getvalueExecCmd)
	rootCmd.AddCommand(setvalueExecCmd)
}
