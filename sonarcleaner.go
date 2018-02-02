package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"sync"
	"encoding/json"
	"strings"
	"time"
	"net/url"
	"errors"
)

type Resource struct {
ID           int    `json:"id"`
UUID         string `json:"uuid"`
Key          string `json:"key"`
Name         string `json:"name"`
Scope        string `json:"scope"`
Qualifier    string `json:"qualifier"`
CreationDate string `json:"creationDate"`
Date         string `json:"date"`
Lname        string `json:"lname"`
Version      string `json:"version"`
Description  string `json:"description,omitempty"`
}

type ProjectResponse []struct {
	Collection []ProjectVersions
}


type ProjectVersions struct {
	ID string `json:"id"`
	Key  string `json:"k"`
	Name string `json:"nm"`
	Sc string `json:"sc"`
	Qualifier string `json:"qu"`
	LatestVersion string `json:"lv"`
	Versions  []Version `json:"v"`
}

type Version struct {
Sid string `json:"sid"`
D   string `json:"d"`
}

type ResourceResponse []struct {
	Collection []Resource
}

var wg sync.WaitGroup

func deleteObjectByUUID(id string,sonarUrl string) {
	//fmt.Println("Deleting " + key)
	client := http.Client{
		//Timeout: timeout,
	}
	fullUrl:= sonarUrl + "/api/projects/delete"

	resp, err := client.PostForm(fullUrl,url.Values{"id": {id}})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func detectIfDeletable(resource Resource)(detelable bool) {
	detelable = false

	resourceUpdateDate, err := time.Parse("2006-01-02T15:04:05.999999999Z0700", string(resource.Date))
	if err != nil {
		panic(err)
	}
	//fmt.Println(resourceUpdateDate)
	if resourceUpdateDate.Before(time.Now().AddDate(0, -1, 0)) {
		if strings.Contains(resource.Key, ":feature_") && !strings.Contains(resource.Key, ":master") && !strings.Contains(resource.Key, ":develop") && !strings.Contains(resource.Key, ":release")  {
			fmt.Println(resource)
			fmt.Printf("Anylysis Date: %s Key: %s is deletable. \n",  string(resource.Date), string(resource.Key))
			detelable = true
		}
	}
return
}

func getProjects(sonarUrl string) (projects []ProjectVersions) {
	client := http.Client{
		//Timeout: timeout,
	}
	fullUrl:= sonarUrl + "/api/projects/index?versions=true"
	resp, err := client.Get(fullUrl)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	rr, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(rr))

	projects = make([]ProjectVersions,0)
	json.Unmarshal(rr, &projects)
	return
}

func getResourceById(sonarUrl string, id string ) (resource Resource, err error) {
	client := http.Client{
		//Timeout: timeout,
	}
	fullUrl:=  sonarUrl + "/api/resources/index?resource=" + id
	resp, err := client.Get(fullUrl)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	rr, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(rr))

	var r []Resource

	json.Unmarshal(rr, &r)
	if(r!=nil && len(r)>=0){
		fmt.Println(r)
		resource = r[0]
	}else{
		fmt.Println("Api Returned Empty Array")
		return resource,errors.New("Api Returned Empty Array")
	}
	return
}


func main() {
	maxProcs := runtime.NumCPU()
	p := fmt.Println
	runtime.GOMAXPROCS(maxProcs)
	sonarUrl:= "https://user:pass@sonar.example.com"

	projects:= getProjects(sonarUrl)
	p(len(projects))

	for _, project := range projects {
		resource,err:= getResourceById(sonarUrl,project.ID)
		if(err==nil && detectIfDeletable(resource)){
			wg.Add(1)
			deleteObjectByUUID(resource.UUID,sonarUrl)
		}
	}

	p("Waiting for all goroutines...")
	wg.Wait()
	p("Done")

}