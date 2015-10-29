package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/jawher/mow.cli"
)

var (
	app = cli.App("apisonator", "")
)

func main() {

	app.Command("register", "Register user to Apisonator.io", register)
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
			resp, err := http.PostForm("http://api.apisonator.io/api/register", data)
			if err != nil {
				panic(err)
			}
			fmt.Println(resp)
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
			resp, err := http.PostForm("http://api.apisonator.io/api/register", data)
			if err != nil {
				panic(err)
			}
			fmt.Println(resp)
		}
	}

}
