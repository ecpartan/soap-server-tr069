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

	"github.com/ecpartan/soap-server-tr069/internal/config"
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
		fmt.Println(arr)

		var result string
		for _, line := range arr {
			result += *line + "\n"
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

		url := "http://localhost:8088/frontcli"

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

}
