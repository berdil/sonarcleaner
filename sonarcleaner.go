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
	fmt.Println("URL:>", fullUrl)

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
	t1, _ := time.Parse("2006-01-02T15:04:05+0000", string(resource.Date))
	if t1.Before(time.Now().AddDate(0, -3, 0)) {
		if strings.Contains(resource.Key, ":feature_") && !strings.Contains(resource.Key, ":master") && !strings.Contains(resource.Key, ":develop") && !strings.Contains(resource.Key, ":release")  {
			fmt.Printf("Project:%s Anylysis Date: %s Key: %s is deletable. \n", string(resource.Name), string(resource.Date), string(resource.Key))
			detelable = true
		}
	}
	return
}

func getResources(sonarUrl string) (resources []Resource) {
	client := http.Client{
		//Timeout: timeout,
	}
	fullUrl:= sonarUrl + "/api/resources/index?metrics=coverage&limit=1000"
	resp, err := client.Get(fullUrl)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	rr, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(rr))

	resources = make([]Resource,0)
	json.Unmarshal(rr, &resources)
	return
}


func main() {
	maxProcs := runtime.NumCPU()
	p := fmt.Println
	runtime.GOMAXPROCS(maxProcs)
	sonarUrl:= "https://user:pass@sonar.example.com"

	resources:= getResources(sonarUrl)
	p(len(resources))

	for _, resource := range resources {
		if(detectIfDeletable(resource)){
			p(resource.UUID)
			wg.Add(1)
			deleteObjectByUUID(resource.UUID,sonarUrl)
		}
	}
	p("Waiting for all goroutines...")
	wg.Wait()
	p("Done")

}
