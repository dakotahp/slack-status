package main

import(
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
  "github.com/spf13/viper"
)

const holidayEmoji string = ":palm_tree:"
const holidayStatus string = "OOO"

const lunchEmoji string = ":sandwich:"
const lunchStatus string = "Out to lunch"

const workEmoji string = ":male-technologist::skin-tone-3:"
const workStatus string = "Hard at work"

const wfhEmoji string = ":house_with_garden:"
const wfhStatus string = "Working from home"

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

func main() {
  var inputWorkspace string
  var action string

  viper.SetConfigName(".slackstatus")
  viper.AddConfigPath(".")
  viper.AddConfigPath("$HOME/")

  err := viper.ReadInConfig()
  if err != nil {
    if _, ok := err.(viper.ConfigFileNotFoundError); ok {
      panic(fmt.Errorf("Fatal error config file: %s \n", err))
    } else {
      panic(fmt.Errorf("Other error with config file: %s \n", err))
    }
  }

  if len(os.Args) == 2 {
    action = os.Args[1]
  } else {
    inputWorkspace = os.Args[1]
    action = os.Args[2]
  }

  loadWorkspaces(inputWorkspace)
  loadStatuses(action)

  SetStatus(selectedStatus.StatusText, selectedStatus.Emoji)
  SetPresence(selectedStatus.Presence)

  // switch {
  //   case action == "offline":
  //     LeavingTime()
  //   case action == "work":
  //     WorkTime()
  //   case action == "wfh":
  //     WorkFromHomeTime()
  //   case action == "ooo":
  //     HolidayTime(os.Args[3])
  //   case action == "lunch":
  //     LunchTime()
  //   case action == "test":
  //     TestAuth()
  //   default:
  //     fmt.Println("Enter offline, work, wfh, lunch");
  // }
}

func loadWorkspaces(inputWorkspace string) {
  configWorkspaces := viper.GetStringSlice("workspaces")

  for i := 0; i < len(configWorkspaces); i++ {
    wsp := Workspace{
      ShortName: viper.GetString("workspace_credentials." + configWorkspaces[i] + ".short_name"),
      Token: viper.GetString("workspace_credentials." + configWorkspaces[i] + ".token"),
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
  configStatuses := viper.GetStringSlice("all_statuses")
  fmt.Println("found statuses", configStatuses)

  for i := 0; i < len(configStatuses); i++ {
    fmt.Println("specified action", action)
    fmt.Println("Looking for corresponding shortname", viper.GetString("statuses." + configStatuses[i] + ".short_name"))
    fmt.Println("Yaml key not found:", "statuses." + configStatuses[i] + ".short_name")
    fmt.Println("\n")

    if viper.GetString("statuses." + configStatuses[i] + ".short_name") == action {
      selectedStatus = Status{
        Emoji: viper.GetString("statuses." + configStatuses[i] + ".emoji"),
        Presence: viper.GetString("statuses." + configStatuses[i] + ".presence"),
        ShortName: viper.GetString("statuses." + configStatuses[i] + ".short_name"),
        StatusText: viper.GetString("statuses." + configStatuses[i] + ".status_text"),
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

func HolidayTime(leaveTerm string) {
  SetStatus(holidayStatus + " " + leaveTerm, holidayEmoji)
  SetPresence("away")
}

func LeavingTime() {
  SetStatus("", "")
  SetPresence("away")
}

func LunchTime() {
  SetStatus(lunchStatus, lunchEmoji)
  SetPresence("away")
}

func WorkTime() {
  SetStatus(workStatus, workEmoji)
  SetPresence("auto")
}

func WorkFromHomeTime() {
  SetStatus(wfhStatus, wfhEmoji)
  SetPresence("auto")
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

  b, err := ioutil.ReadAll(resp.Body)
  fmt.Println(string(b))

  resp.Body.Close()

  if err != nil {
    fmt.Fprintf(os.Stderr, "fetch: reading %s: %v\n", err)
    os.Exit(1)
  }
}
