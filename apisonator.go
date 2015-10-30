package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"strconv"

	"github.com/ttacon/chalk"
	"gopkg.in/yaml.v2"

	"github.com/kyokomi/emoji"

	"github.com/jawher/mow.cli"
)

var (
	//APIEndpoint = "http://07225c10.ngrok.io"
	APIEndpoint = "http://api.apisonator.io"
)

type loginResponse struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	APIKey    string `json:"api_key"`
}

type ReleaseResponse struct {
	ID        int    `json:"id"`
	Version   string `json:"version"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Config    string `json:"config"`
	ProxyID   int    `json:"proxy_id"`
}

//
// {"id":17,"user_id":24,"version":"c1Svj9kv1XQQcr9id8Q3WTJH","created_at":"2015-10-29T23:29:31.225Z","updated_at":"2015-10-29T23:29:31.225Z","config":"subdomain: miaumiau1\nmiddleware:\n- middleware01\n- middleware02\n","proxy_id":50}

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

	cmd.Spec = "EMAIL PASSWORD"

	var (
		email    = cmd.StringArg("EMAIL", "", "Your email")
		password = cmd.StringArg("PASSWORD", "", "Your password")
	)

	cmd.Action = func() {

		if *email != "" && *password != "" {
			data := url.Values{}
			data.Set("email", *email)
			data.Add("password", *password)
			resp, err := http.PostForm(APIEndpoint+"/api/registrations.json", data)
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

	cmd.Spec = "EMAIL PASSWORD"

	var (
		email    = cmd.StringArg("EMAIL", "", "Your email")
		password = cmd.StringArg("PASSWORD", "", "Your password")
	)

	cmd.Action = func() {

		if *email != "" && *password != "" {
			data := url.Values{}
			data.Set("email", *email)
			data.Add("password", *password)
			resp, err := http.PostForm(APIEndpoint+"/api/sessions.json", data)
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

		resp, _ := http.PostForm(APIEndpoint+"/api/proxies.json", data)
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
	cmd.Spec = "[--config-path=<dir>]"

	var (
		bootstrapPath = cmd.StringOpt("config-path", "./", "Parent directory for your config.yml")
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
		fyml, _ := ioutil.ReadFile(*bootstrapPath + "/config.yml")
		data := url.Values{}
		data.Set("api_key", apiKey)
		data.Add("config", string(fyml))
		resp, _ := http.PostForm(APIEndpoint+"/api/releases.json", data)

		fmt.Println("\nUpdated configuration\n")
		var response ReleaseResponse
		body, _ := ioutil.ReadAll(resp.Body)

		if err := json.Unmarshal(body, &response); err != nil {
			panic(err)
		}

		yamlFile, err := ioutil.ReadFile(*bootstrapPath + "/config.yml")
		if err != nil {
			fmt.Println("No config.yml found in dir, use --config-path=<dir> and point to parent dir.")
			os.Exit(1)
		}
		var config configYaml
		err = yaml.Unmarshal(yamlFile, &config)
		if err != nil {
			panic(err)
		}
		fmt.Println("\nDeploying middlewares: ")

		for _, middleware := range config.Middleware {

			middlewareFile, err := ioutil.ReadFile(*bootstrapPath + "/middleware/" + string(middleware) + ".lua")

			if err != nil {
				panic(err)
			}

			fmt.Printf("\t- %s", middleware)
			dataFiles := url.Values{}
			dataFiles.Set("api_key", apiKey)
			dataFiles.Add("release_id", strconv.Itoa(response.ID))
			dataFiles.Add("name", string(middleware))
			dataFiles.Add("content", string(middlewareFile))
			//			fmt.Println(dataFiles)
			resp, _ := http.PostForm(APIEndpoint+"/api/functions.json", dataFiles)
			if resp.StatusCode != http.StatusCreated {
				fmt.Println("\nSomething went wrong.. are middlewares specified correctly?")
				os.Exit(1)
			} else {
				fmt.Printf(" OK.\n")
			}
		}

		data = url.Values{}
		data.Set("api_key", apiKey)
		data.Add("release_id", strconv.Itoa(response.ID))
		data.Add("done", "true")
		r := data.Encode()
		client := &http.Client{}
		req, err := http.NewRequest("PUT", APIEndpoint+"/api/releases/"+strconv.Itoa(response.ID)+".json",
			bytes.NewBufferString(r))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
		resp, err = client.Do(req)
		defer resp.Body.Close()
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != http.StatusNoContent {
			fmt.Println("Something went wrong :(")
		} else {
			fmt.Println("\nDeployed!\n")
		}
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
