package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/jawher/mow.cli"
)

type loginResponse struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	APIKey    string `json:"api_key"`
}

var (
	app = cli.App("apisonator", "")
)

func main() {

	app.Command("register", "Register user to Apisonator.io", register)
	app.Command("login", "Login to Apisonator.io", login)

	app.Run(os.Args)

}

func register(cmd *cli.Cmd) {

	cmd.Spec = "--email=<email> --password=<password>"

	var (
		email    = cmd.StringOpt("email", "", "Your email")
		password = cmd.StringOpt("password", "", "Your password")
	)

	cmd.Action = func() {

		if *email != "" && *password != "" {
			data := url.Values{}
			data.Set("email", *email)
			data.Add("password", *password)
			resp, err := http.PostForm("http://api.apisonator.io/api/registrations.json", data)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)

			if resp.StatusCode == http.StatusCreated {

				// Extract this shit
				var response loginResponse
				if err := json.Unmarshal(body, &response); err != nil {
					panic(err)
				}
				//Is always scary to remove things..
				authFilePath := "/tmp/cacafuti"
				authFilePath = os.Getenv("HOME") + "/.apisonator"
				os.Remove(authFilePath)
				authFile, _ := os.Create(authFilePath)
				defer authFile.Close()
				fmt.Fprintln(authFile, response.APIKey)

				fmt.Println("Registered correctly. Logged!")

			} else {
				fmt.Println("Invalid Email")
			}
		}
	}
}

func login(cmd *cli.Cmd) {

	cmd.Spec = "--email=<email> --password=<password>"

	var (
		email    = cmd.StringOpt("email", "", "Your email")
		password = cmd.StringOpt("password", "", "Your password")
	)

	cmd.Action = func() {

		if *email != "" && *password != "" {
			data := url.Values{}
			data.Set("email", *email)
			data.Add("password", *password)
			resp, err := http.PostForm("http://api.apisonator.io/api/sessions.json", data)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)

			if resp.StatusCode == http.StatusCreated {

				// Extract this shit
				var response loginResponse
				if err := json.Unmarshal(body, &response); err != nil {
					panic(err)
				}
				//Is always scary to remove things..
				authFilePath := "/tmp/cacafuti"
				authFilePath = os.Getenv("HOME") + "/.apisonator"
				os.Remove(authFilePath)
				authFile, _ := os.Create(authFilePath)
				defer authFile.Close()
				fmt.Fprintln(authFile, response.APIKey)

				fmt.Println("Logged!")

			} else {
				fmt.Println("Invalid User/Password")
			}
		}
	}
}
