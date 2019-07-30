/*
Copyright © 2019 Dakotah Pena <dakotah@pena.name>

*/
package cmd

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
  "strings"

  "github.com/spf13/cobra"
  homedir "github.com/mitchellh/go-homedir"
  "github.com/spf13/viper"
)


var cfgFile string

const baseUri string = "https://slack.com/api"
const presenceUri string = "/users.setPresence"
const statusUri string = "/users.profile.set"

var selectedWorkspace Workspace
var selectedStatus Status

//
// Config Schemas
//

type Status struct {
  Emoji string `yaml:"emoji"`
  Presence string `yaml:"presence"`
  ShortName string `yaml:"short_name"`
  StatusText string `yaml:"status"`
}

var allStatuses []Status

type Workspace struct {
  ShortName string `yaml:"short_name"`
  Token string `yaml:"token"`
}

var allWorkspaces []Workspace

//
// JSON Payloads
//
type PresencePayload struct {
  Presence string `json:"presence"`
}

type ProfilePayload struct {
  StatusText string `json:"status_text"`
  StatusEmoji string `json:"status_emoji"`
}

type StatusPayload struct {
  Profile ProfilePayload `json:"profile"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
  Use:   "slackstatus [workspace] status",
  Short: "Update your slack status quickly from the command line.",
  Long: `Use to update your emoji and status text in any number of Slack workspaces. Use default
commands or customize with your own.

For example, if you're going to lunch you can enter the following:

  $ slackstatus lunch

It will update all Slack workspaces if your config file. To target just one of them if it was set to "work", enter:

  $ slackstatus work lunch

Default statuses available to use:
- work (sets you back to active and clears out status text)
- lunch
- wfh (working from home)
- ooo (out of office)
  `,
  Args: cobra.RangeArgs(1, 2),
  Run: func(cmd *cobra.Command, args []string) {
    var inputWorkspace string
    var action string

    if len(args) == 1 {
      action = args[0]
    } else {
      inputWorkspace = args[0]
      action = args[1]
    }

    loadWorkspaces(inputWorkspace)
    loadStatuses(action)

    SetStatus(selectedStatus.StatusText, selectedStatus.Emoji)
    SetPresence(selectedStatus.Presence)
  },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
  if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func init() {
  cobra.OnInitialize(initConfig)

  // Here you will define your flags and configuration settings.
  // Cobra supports persistent flags, which, if defined here,
  // will be global for your application.

  rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.slackstatus.yaml)")
}


// initConfig reads in config file and ENV variables if set.
func initConfig() {
  if cfgFile != "" {
    // Use config file from the flag.
    viper.SetConfigFile(cfgFile)
  } else {
    // Find home directory.
    home, err := homedir.Dir()
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }

    // Search config in home directory with name ".slackstatus" (without extension).
    viper.AddConfigPath(home)
    viper.SetConfigName(".slackstatus")
  }

  viper.AutomaticEnv() // read in environment variables that match

  // If a config file is found, read it in.
  if err := viper.ReadInConfig(); err == nil {
    fmt.Println("Using config file:", viper.ConfigFileUsed())
  } else {
    fmt.Println("⚠️  No config file found. Copy the example config file from the repo to your home directory:")
    fmt.Println("  $ cp .example.slackstatus.yml ~/.slackstatus.yml\n")
  }
}


func loadWorkspaces(inputWorkspace string) {
  configWorkspaces := viper.GetStringMapString("workspace_credentials")

  for k, _ := range configWorkspaces {
    workspaceName := strings.TrimSuffix(k, ":")

    wsp := Workspace{
      ShortName: viper.GetString("workspace_credentials." + workspaceName + ".short_name"),
      Token: viper.GetString("workspace_credentials." + workspaceName + ".token"),
    }

    allWorkspaces = append(allWorkspaces, wsp)
  }

  if inputWorkspace != "" {
    for _, wsp := range allWorkspaces {
      if wsp.ShortName == inputWorkspace {
        selectedWorkspace = wsp
      }
    }
  }
}

func loadStatuses(action string) {
  configStatuses := viper.GetStringMapString("statuses")

  for k, _ := range configStatuses {
    statusName := strings.TrimSuffix(k, ":")

    if viper.GetString("statuses." + statusName + ".short_name") == action {
      selectedStatus = Status{
        Emoji: viper.GetString("statuses." + statusName + ".emoji"),
        Presence: viper.GetString("statuses." + statusName + ".presence"),
        ShortName: viper.GetString("statuses." + statusName + ".short_name"),
        StatusText: viper.GetString("statuses." + statusName + ".status_text"),
      }
    }

    allStatuses = append(allStatuses, selectedStatus)
  }

  if (Status{}) == selectedStatus {
    fmt.Println("Status not found:", action)
    os.Exit(1)
  }
}

func TestAuth() {
  client := &http.Client {}

  req, err := http.NewRequest("GET", baseUri + "/auth.test", nil)
  if err != nil {
    fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
    os.Exit(1)
  }

  req.Header.Add("content-type", "application/json")
  req.Header.Add("Authorization", "Bearer " + selectedWorkspace.Token)

  resp, err := client.Do(req)

  b, err := ioutil.ReadAll(resp.Body)

  resp.Body.Close()
  if err != nil {
    fmt.Fprintf(os.Stderr, "fetch: reading %s: %v\n", err)
    os.Exit(1)
  }

  fmt.Printf("%s", b)
}

func SetStatus(statusText string, statusEmoji string) {
  profileData := ProfilePayload{
    StatusText: statusText,
    StatusEmoji: statusEmoji,
  }

  statusData := StatusPayload{ Profile: profileData }

  if (Workspace{}) == selectedWorkspace {
    for _, wsp := range allWorkspaces {
      SendRequest(statusData, statusUri, wsp)
    }
  } else {
    SendRequest(statusData, statusUri, selectedWorkspace)
  }
}

func SetPresence(presence string) {
  presenceData := PresencePayload{ Presence: presence }

  if (Workspace{}) == selectedWorkspace {
    for _, wsp := range allWorkspaces {
      SendRequest(presenceData, presenceUri, wsp)
    }
  } else {
    SendRequest(presenceData, presenceUri, selectedWorkspace)
  }
}

func SendRequest(payload interface{}, uri string, wsp Workspace) {
  client := &http.Client {}

  jsonPayload, err := json.Marshal(payload)

  if err != nil {
    fmt.Println(err)
    return
  }

  req, err := http.NewRequest("POST", baseUri + uri, bytes.NewBuffer(jsonPayload))

  if err != nil {
    fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
    os.Exit(1)
  }

  req.Header.Add("content-type", "application/json; charset=UTF-8")
  req.Header.Add("Authorization", "Bearer " + wsp.Token)

  resp, err := client.Do(req)

  if err == nil {
    fmt.Println("Request successful!")
  }

  // b, err := ioutil.ReadAll(resp.Body)
  // fmt.Println(string(b))

  resp.Body.Close()

  if err != nil {
    fmt.Fprintf(os.Stderr, "fetch: reading %s: %v\n", err)
    os.Exit(1)
  }
}
