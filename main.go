package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ak1ra24/tfstatediff/ci"
	"github.com/ak1ra24/tfstatediff/githubapi"
	"gopkg.in/yaml.v2"
)

type Tfstate struct {
	Resources []Resource `json:"resources"`
}

type Resource struct {
	Module    string     `json:"module"`
	Mode      string     `json:"mode"`
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Instances []Instance `json:"instances"`
}

type Instance struct {
	Schema     int       `json:"schema_version"`
	Attributes Attribute `json:"attributes"`
}

type Attribute struct {
	Name      string `json:"name"`
	Ipaddress string `json:"ipaddress"`
}

type NotifierService struct {
	Ci       string   `yaml:"ci"`
	Notifier Notifier `yaml:"notifier"`
}

type Notifier struct {
	Github Github `yaml:"github"`
}

type Github struct {
	Token      string `yaml:"token"`
	Repository struct {
		Owner string `yaml:"owner"`
		Repo  string `yaml:"name"`
	} `yaml:"repository"`
}

func ReadYaml(filename string) NotifierService {
	// yamlを読み込む
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var notifier NotifierService
	err = yaml.Unmarshal(buf, &notifier)
	if err != nil {
		panic(err)
	}
	return notifier
}

func tfstatediff(oldtfstate, newtfstate string) (string, error) {
	newjson, err := ioutil.ReadFile(newtfstate)
	if err != nil {
		return "", err
	}

	var newResources Tfstate
	if err := json.Unmarshal(newjson, &newResources); err != nil {
		return "", err
	}

	newServer := map[string]string{}
	for _, r := range newResources.Resources {
		if r.Type == "sakuracloud_server" {
			for _, instance := range r.Instances {
				newServer[instance.Attributes.Name] = instance.Attributes.Ipaddress
			}
		}
	}
	// fmt.Println(newServer)

	oldjson, err := ioutil.ReadFile(oldtfstate)
	if err != nil {
		return "", err
	}

	var oldResources Tfstate
	if err := json.Unmarshal(oldjson, &oldResources); err != nil {
		return "", err
	}

	oldServer := map[string]string{}
	for _, r := range oldResources.Resources {
		if r.Type == "sakuracloud_server" {
			for _, instance := range r.Instances {
				oldServer[instance.Attributes.Name] = instance.Attributes.Ipaddress
			}
		}
	}

	if len(newServer) != len(oldServer) {
		for newkey, _ := range newServer {
			if _, ok := oldServer[newkey]; ok {
				delete(newServer, newkey)
			}
		}
	}

	output, err := json.Marshal(newServer)
	if err != nil {
		log.Fatal(err)
	}

	return string(output), nil
}

func main() {
	// commands args
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Set config yaml")
		os.Exit(1)
	}

	// flag option
	var (
		oldstate string
		newstate string
	)
	flag.StringVar(&oldstate, "old", "", "old tfstate file")
	flag.StringVar(&newstate, "new", "", "new tfstate file")
	flag.Parse()

	if oldstate != "" && newstate != "" {
		// tfstate diff
		outputs, err := tfstatediff(oldstate, newstate)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(outputs)

		// github pr comment
		notifier := ReadYaml(args[1])
		ciname := notifier.Ci
		github_settings := notifier.Notifier.Github

		var ciservice ci.CiService
		switch ciname {
		case "drone":
			ciservice, err = ci.Drone()
			if err != nil {
				panic(err)
			}
		case "":
			fmt.Errorf("Set CI Service")
		default:
			fmt.Errorf("Not Support")
		}

		pr := ciservice.PR

		client := githubapi.NewClient(github_settings.Repository.Owner, github_settings.Repository.Repo, github_settings.Token, pr)
		if ciservice.Event == "pull_request" {
			if err := client.PRComment(outputs); err != nil {
				log.Fatal(err)
			}
		} else if ciservice.Event == "push" && ciservice.Branch == "master" {
			if err := client.PRComment(outputs); err != nil {
				log.Fatal(err)
			}
		}

	} else {
		fmt.Println("Not Set tfstate file")
		os.Exit(1)
	}
}
