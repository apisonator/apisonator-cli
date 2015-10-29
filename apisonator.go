package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ttacon/chalk"
	"gopkg.in/yaml.v2"

	"github.com/kyokomi/emoji"

	"github.com/jawher/mow.cli"
)

type loginResponse struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	APIKey    string `json:"api_key"`
}

type configYaml struct {
	Subdomain  string   `yaml:"subdomain"`
	Middleware []string `yaml:"middleware"`
}

type createEndpoint struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Endpoint  string `json:"endpoint"`
	Subdomain string `json:"subdomain"`
}

var (
	app = cli.App("apisonator", "")
)

func main() {

	app.Command("register", "Register user to Apisonator.io", register)
	app.Command("login", "Login to Apisonator.io", login)
	app.Command("create", "Create your apisonator endpoint", create)
	app.Command("deploy", "Deploy your apisonator endpoint", deploy)

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

func create(cmd *cli.Cmd) {

	cmd.Spec = "SUBDOMAIN ENDPOINT [--no-bootstrap | --bootstrap-destination=<dir>]"

	var (
		name          = cmd.StringArg("SUBDOMAIN", "", "Name for your apisonator proxy $subdomain.apisonator.io")
		endpoint      = cmd.StringArg("ENDPOINT", "", "Your API endpoint")
		noBootstrap   = cmd.BoolOpt("no-bootstrap", false, "Don't create the basic app bootstrap")
		bootstrapPath = cmd.StringOpt("bootstrap-destination", "./", "Path to create the bootstrap files for your project")
	)

	cmd.Action = func() {
		var apiKey string
		authFilePath := os.Getenv("HOME") + "/.apisonator"
		f, err := os.Open(authFilePath)
		if err != nil {
			fmt.Println("Error. Login first")
			os.Exit(1)
		}
		fmt.Fscan(f, &apiKey)

		data := url.Values{}
		data.Set("subdomain", *name)
		data.Add("endpoint", *endpoint)
		data.Add("api_key", apiKey)

		resp, _ := http.PostForm("http://api.apisonator.io/api/proxies.json", data)
		body, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusCreated {
			var response createEndpoint
			if err := json.Unmarshal(body, &response); err != nil {
				panic(err)
			}
			Success := emoji.Sprintf("\n:white_check_mark: Your apisonator endpoint: %shttp://%s.apisonator.io%s -> %s\n", chalk.Green, response.Subdomain, chalk.Reset, response.Endpoint)
			fmt.Print(Success)

			if *noBootstrap {
				fmt.Println("\nBootstrap not created\n")
			} else {

				// Ugly eh. Extract
				resp, _ := http.Get("https://github.com/apisonator/bootstrap/archive/master.zip")
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)
				mode := int(0777)
				os.Remove("/tmp/bootstrap.zip")
				ioutil.WriteFile("/tmp/bootstrap.zip", body, os.FileMode(mode))
				Unzip("/tmp/bootstrap.zip", *bootstrapPath)
				os.Rename(*bootstrapPath+"bootstrap-master", *bootstrapPath+"apisonator-"+*name)
				fmt.Printf("\t\nBootstrap directory created at: %s\n\n", *bootstrapPath+"apisonator-"+*name)

				// Modify yaml file and set correct subdomain
				yamlFile, err := ioutil.ReadFile(*bootstrapPath + "apisonator-" + *name + "/config.yml")
				var config configYaml
				err = yaml.Unmarshal(yamlFile, &config)
				if err != nil {
					panic(err)

				}
				config.Subdomain = *name
				mary, err := yaml.Marshal(config)
				if err != nil {
					panic(err)

				}
				err = ioutil.WriteFile(*bootstrapPath+"apisonator-"+*name+"/config.yml", mary, os.FileMode(mode))
				if err != nil {
					panic(err)

				}
			}

		} else {
			// move this emoji and success / fail to another
			Failed := emoji.Sprintf("\n:red_circle: Subdomain %s does exists\n", *name)
			fmt.Println(Failed)
		}
	}
}

func deploy(cmd *cli.Cmd) {
	cmd.Spec = ""

	var ()

	cmd.Action = func() {

		var apiKey string
		authFilePath := os.Getenv("HOME") + "/.apisonator"
		f, err := os.Open(authFilePath)

		if err != nil {
			fmt.Println("Error. Login first")
			os.Exit(1)
		}

		fmt.Fscan(f, &apiKey)
		fyml, _ := ioutil.ReadFile("./test.yml")
		data := url.Values{}
		data.Set("api_key", apiKey)
		data.Add("config", string(fyml))
		resp, _ := http.PostForm("http://api.apisonator.io/api/releases.json", data)
		fmt.Println(resp)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
	}
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
