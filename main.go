package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	// "github.com/barbich/restish-swagger/swagger"
	"github.com/barbich/upguard-cli/swagger"
	"github.com/barbich/upguard-cli/upguard"
	"github.com/danielgtaylor/restish/cli"
	log "github.com/sirupsen/logrus"
	// "github.com/google/martian/log"
)

var version string = "dev"
var myName string = "upguard-cli"

var defaultConfig = []byte(`{
    "$schema": "https://rest.sh/schemas/apis.json",
    "upguard-cli": {
	  "base": "https://cyber-risk.upguard.com/api/public",
    //   "operation_base": "/api/public",
	  "spec_files": [
        "https://cyber-risk.upguard.com/api/swagger.json"
      ],
	  "profiles": {
        "default": {}
      }
    }
  }
`)

func init() {
	if len(os.Args) > 1 {
		if os.Args[1] != myName {
			os.Args = append([]string{os.Args[0], myName}, os.Args[1:]...)
		}
	}

	// Check if a config file is existing otherwise create and populate ...
	configDir := getConfigDir(myName)
	log.Debug("configDir:", configDir)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		panic(err)
	}

	configFile := filepath.Join(configDir, "apis.json")

	if _, err := os.Stat(configFile); err != nil {
		if err := os.WriteFile(configFile, defaultConfig, 0666); err != nil {
			log.Fatal(err)
		}
	}

	// add required --rsh-header from ENV
	headers, found := os.LookupEnv(strings.ToUpper(strings.Replace(myName, "-", "_", -1)) + "_UPGUARDKEY")
	if !found {
		var reAuthorization = regexp.MustCompile(`(?i)--rsh-header.authorization:(\w+)`)
		if !reAuthorization.MatchString(strings.Join(os.Args, " ")) {
			log.Info(fmt.Sprintf("Key %s is required\n", strings.ToUpper(strings.Replace(myName, "-", "_", -1))+"_UPGUARDKEY"))
			log.Info("Alternative: set parameter --rsh-header authorization:XXXXX where XXXXX is the Upguard Key")
		}
	} else {
		log.Debug(fmt.Sprintf("Key %s Found\n", strings.ToUpper(strings.Replace(myName, "-", "_", -1))+"_UPGUARDKEY"))
		os.Args = append(os.Args[0:], "--rsh-header", fmt.Sprintf("authorization:%s", headers))
	}
}

func getConfigDir(appName string) string {
	configDirEnv := strings.ToUpper(appName) + "_CONFIG_DIR"
	configDir := os.Getenv(configDirEnv)

	if configDir == "" {
		// Create new config directory
		configBase, _ := os.UserConfigDir()
		configDir = filepath.Join(configBase, appName)
	}
	return configDir
}

func main() {
	if version == "dev" {
		// Try to add the executable modification time to the dev version.
		filename, _ := os.Executable()
		if info, err := os.Stat(filename); err == nil {
			version += "-" + info.ModTime().Format("2006-01-02-15:04")
		}
	}
	cli.Init(myName, version)
	// fmt.Println(viper.Get("app-name"))

	// Ensure Root.Aliases is being populated
	if len(cli.Root.Aliases) == 0 {
		cli.Root.Aliases = append(cli.Root.Aliases, myName)
	}

	// Register default encodings, content type handlers, and link parsers.
	cli.Defaults()

	// bulk.Init(cli.Root)

	// Register format loaders to auto-discover API descriptions
	cli.AddLoader(swagger.New())
	// cli.AddLoader(openapi.New())

	// Add LinkParser for UpGuard
	cli.AddLinkParser(&upguard.UpguardAPIParser{})

	// See: https://github.com/spf13/cobra/issues/725
	// Forcing the command to be only myName
	if len(os.Args) <= 1 {
		fmt.Println("No sufficient arguments.")
		os.Args = append([]string{os.Args[0], cli.Root.Aliases[0]}, os.Args[1:]...)
	} else {
		if os.Args[1] != myName {
			os.Args = append([]string{os.Args[0], cli.Root.Aliases[0]}, os.Args[1:]...)
		}
	}

	// Run the CLI, parsing arguments, making requests, and printing responses.
	if err := cli.Run(); err != nil {
		os.Exit(1)
	}

	// Exit based on the status code of the last request.
	os.Exit(cli.GetExitCode())
}
