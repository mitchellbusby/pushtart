package main

import (
	"fmt"
	"io"
	"path"
	"pushtart/config"
	"pushtart/tartmanager"
	"strconv"
	"strings"
)

func listTarts(params map[string]string, w io.Writer, user string) {
	for pushURL, tart := range config.All().Tarts {
		fmt.Fprint(w, tart.Name+" ("+pushURL+"): ")
		if tart.IsRunning {
			fmt.Fprint(w, "Running (PID "+strconv.Itoa(tart.PID)+") ")
		} else {
			fmt.Fprint(w, "Stopped. ")
		}

		if tart.LogStdout {
			fmt.Fprintln(w, "[Stdout -> Log is ENABLED]")
		} else {
			fmt.Fprintln(w, "[Stdout -> Log is disabled]")
		}

		if len(tart.Env) > 0 {
			for _, env := range tart.Env {
				fmt.Fprintln(w, "\t"+env)
			}
		}
	}
}

func newTart(params map[string]string, w io.Writer, user string) {
	if missingFields := checkHasFields([]string{"tart"}, params); len(missingFields) > 0 {
		fmt.Fprintln(w, "USAGE: pushtart new-tart --tart <pushURL>")
		printMissingFields(missingFields, w)
		return
	}

	exists, _ := findTart(params["tart"])
	if exists {
		fmt.Fprintln(w, "Err: A tart by that pushURL already exists")
		return
	}
	if user == "" {
		fmt.Fprintln(w, "Err: New tarts can only be presubmitted (created) from the management console.")
		return
	}
	if !strings.HasPrefix(params["tart"], "/") {
		fmt.Fprintln(w, "Err: pushURLs must start with a '/' character.")
		return
	}

	err := tartmanager.PreGitRecieve(params["tart"], user)
	if err != nil {
		fmt.Fprintln(w, "Err:", err)
	}
	tartmanager.New(params["tart"], user)
}

func digestTartConfig(params map[string]string, w io.Writer, user string) {
	if missingFields := checkHasFields([]string{"tart"}, params); len(missingFields) > 0 {
		fmt.Fprintln(w, "USAGE: pushtart digest-tartconfig --tart <pushURL>")
		printMissingFields(missingFields, w)
		return
	}

	exists, tart := findTart(params["tart"])
	if !exists {
		fmt.Fprintln(w, "Err: A tart by that pushURL does not exist")
		return
	}
	if user != "" && !tartmanager.UserHasTartOwnership(user, tart.Owners) {
		fmt.Fprintln(w, "Err: You ("+user+") are not an owner of the specified tart")
		return
	}
	err := tartmanager.ExecuteCommandFile(path.Join(config.All().DeploymentPath, tart.PushURL, "tartconfig"), tart.PushURL, &w)
	if err != nil {
		fmt.Fprintln(w, "Err:", err)
	}
}

func startTart(params map[string]string, w io.Writer, user string) {
	if missingFields := checkHasFields([]string{"tart"}, params); len(missingFields) > 0 {
		fmt.Fprintln(w, "USAGE: pushtart start-tart --tart <pushURL>")
		printMissingFields(missingFields, w)
		return
	}

	exists, tart := findTart(params["tart"])
	if !exists {
		fmt.Fprintln(w, "Err: A tart by that pushURL does not exist")
		return
	}
	if user != "" && !tartmanager.UserHasTartOwnership(user, tart.Owners) {
		fmt.Fprintln(w, "Err: You ("+user+") are not an owner of the specified tart")
		return
	}
	err := tartmanager.Start(tart.PushURL)
	if err != nil {
		fmt.Fprintln(w, "Err:", err)
	}
}

func stopTart(params map[string]string, w io.Writer, user string) {
	if missingFields := checkHasFields([]string{"tart"}, params); len(missingFields) > 0 {
		fmt.Fprintln(w, "USAGE: pushtart start-tart --tart <pushURL>")
		printMissingFields(missingFields, w)
		return
	}

	exists, tart := findTart(params["tart"])
	if !exists {
		fmt.Fprintln(w, "Err: A tart by that pushURL does not exist")
		return
	}
	if user != "" && !tartmanager.UserHasTartOwnership(user, tart.Owners) {
		fmt.Fprintln(w, "Err: You ("+user+") are not an owner of the specified tart")
		return
	}
	err := tartmanager.Stop(tart.PushURL)
	if err != nil {
		fmt.Fprintln(w, "Err:", err)
	}
}

func findTart(tartName string) (bool, config.Tart) {
	if tartmanager.Exists(tartName) {
		return true, tartmanager.Get(tartName)
	}
	if tartmanager.Exists("/" + tartName) {
		return true, tartmanager.Get("/" + tartName)
	}
	return false, config.Tart{}
}

func editTart(params map[string]string, w io.Writer, user string) {
	if missingFields := checkHasFields([]string{"tart"}, params); len(missingFields) > 0 {
		fmt.Fprintln(w, "USAGE: pushtart edit-tart --tart <pushURL> [--name <name>] [--set-env \"<env-name>=<env-value>\"] [--delete-env <env-name>] [--log-stdout yes/no]")
		printMissingFields(missingFields, w)
		return
	}

	exists, tart := findTart(params["tart"])
	if !exists {
		fmt.Fprintln(w, "Err: A tart by that pushURL does not exist")
		return
	}

	if user != "" && !tartmanager.UserHasTartOwnership(user, tart.Owners) {
		fmt.Fprintln(w, "Err: You ("+user+") are not an owner of the specified tart")
		return
	}

	if params["name"] != "" {
		tart.Name = params["name"]
	}

	if params["set-env"] != "" {
		tart.Env = setEnv(tart.Env, params["set-env"], "")
	}

	if params["delete-env"] != "" {
		tart.Env = setEnv(tart.Env, "", params["delete-env"])
	}

	if params["log-stdout"] != "" {
		if strings.ToLower(params["log-stdout"]) == "yes" {
			tart.LogStdout = true
		} else {
			tart.LogStdout = false
		}
	}

	tartmanager.Save(tart.PushURL, tart)
}

func tartRestartMode(params map[string]string, w io.Writer, user string) {
	if missingFields := checkHasFields([]string{"tart", "enabled"}, params); len(missingFields) > 0 {
		fmt.Fprintln(w, "USAGE: pushtart tart-restart-mode --tart <pushURL> --enabled yes/no [--lull-period <seconds>]")
		printMissingFields(missingFields, w)
		return
	}

	exists, tart := findTart(params["tart"])
	if !exists {
		fmt.Fprintln(w, "Err: A tart by that pushURL does not exist")
		return
	}
	if user != "" && !tartmanager.UserHasTartOwnership(user, tart.Owners) {
		fmt.Fprintln(w, "Err: You ("+user+") are not an owner of the specified tart")
		return
	}

	if strings.ToLower(params["enabled"]) == "yes" {
		tart.RestartOnStop = true
	} else {
		tart.RestartOnStop = false
	}

	if params["lull-period"] != "" {
		i, err := strconv.Atoi(params["lull-period"])
		if err != nil {
			fmt.Fprintln(w, "Err: could not read value for lull-period. Did you provide an integer?")
			fmt.Fprintln(w, "Aborting.")
			return
		}
		tart.RestartDelaySecs = i
	}

	tartmanager.Save(tart.PushURL, tart)
}

func setEnv(envList []string, envString, delString string) []string {
	key := strings.Split(envString, "=")[0]
	var output []string

	for _, envEntry := range envList {
		if (strings.Split(envEntry, "=")[0] == key) || strings.Split(envEntry, "=")[0] == delString {
			//no op
		} else {
			output = append(output, envEntry)
		}
	}
	if envString != "" {
		output = append(output, envString)
	}
	return output
}

func tartAddOwner(params map[string]string, w io.Writer, user string) {
	if missingFields := checkHasFields([]string{"tart", "username"}, params); len(missingFields) > 0 {
		fmt.Fprintln(w, "USAGE: pushtart tart-add-owner --tart <pushURL> --username <username>")
		printMissingFields(missingFields, w)
		return
	}

	exists, tart := findTart(params["tart"])
	if !exists {
		fmt.Fprintln(w, "Err: A tart by that pushURL does not exist")
		return
	}
	if user != "" && !tartmanager.UserHasTartOwnership(user, tart.Owners) {
		fmt.Fprintln(w, "Err: You ("+user+") are not an owner of the specified tart")
		return
	}

	for _, owner := range tart.Owners {
		if owner == params["username"] {
			fmt.Fprintln(w, "Err: "+params["username"]+" is already set as an owner.")
			return
		}
	}

	tart.Owners = append(tart.Owners, params["username"])
	tartmanager.Save(tart.PushURL, tart)
}

func tartRemoveOwner(params map[string]string, w io.Writer, user string) {
	if missingFields := checkHasFields([]string{"tart", "username"}, params); len(missingFields) > 0 {
		fmt.Fprintln(w, "USAGE: pushtart tart-remove-owner --tart <pushURL> --username <username>")
		printMissingFields(missingFields, w)
		return
	}

	exists, tart := findTart(params["tart"])
	if !exists {
		fmt.Fprintln(w, "Err: A tart by that pushURL does not exist")
		return
	}
	if user != "" && !tartmanager.UserHasTartOwnership(user, tart.Owners) {
		fmt.Fprintln(w, "Err: You ("+user+") are not an owner of the specified tart")
		return
	}

	didFind := false
	temp := []string{}
	for _, owner := range tart.Owners {
		if owner == params["username"] {
			didFind = true
		} else {
			temp = append(temp, owner)
		}
	}

	if !didFind {
		fmt.Fprintln(w, "Err: That user is not a tart owner.")
		return
	}

	if len(temp) == 0 {
		fmt.Fprintln(w, "Err: A tart must always have at least own owner. Add another owner or delete the tart.")
		return
	}

	tart.Owners = temp
	tartmanager.Save(tart.PushURL, tart)
}
