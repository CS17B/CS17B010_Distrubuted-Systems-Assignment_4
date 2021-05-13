package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rocketlaunchr/dataframe-go/imports"
	"github.com/thedevsaddam/renderer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var rnd *renderer.Render

type Fact struct {
	Job_Title    string `json:"Job_Title"`
	Company_Name string `json:"Company_Name"`
	Location     string `json:"Location"`
	Start_Date   string `json:"Start_Date"`
	Duration     string `json:"Duration"`
	Stipend      string `json:"Stipend"`
	Last_Date    string `json:"Last_Date"`
}

func init() {
	opts := renderer.Options{
		ParseGlobPattern: "./tpl/*.html",
	}
	rnd = renderer.New(opts)
}

func newedithandler(w http.ResponseWriter, r *http.Request) {
	rnd.HTML(w, http.StatusOK, "newedit", nil)
}

func viewhandler(w http.ResponseWriter, r *http.Request) {

	n := r.FormValue("o")
	if n == "1" {
		n = "Job_Title"
	} else if n == "2" {
		n = "Company_Name"
	} else if n == "3" {
		n = "Location"
	} else if n == "4" {
		n = "Duration"
	} else {
		n = "Stipend"
	}

	t := r.FormValue("n")
	fmt.Println(n)
	fmt.Println(t)

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://teja1510:teja%401510@cluster0.j2grx.mongodb.net/admin"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	quickstartDatabase := client.Database("Internship_Details")
	episodesCollection := quickstartDatabase.Collection("Info")
	filterCursor, err := episodesCollection.Find(ctx, bson.M{n: t})
	if err != nil {
		log.Fatal(err)
	}
	var episodesFiltered []bson.M
	if err = filterCursor.All(ctx, &episodesFiltered); err != nil {
		log.Fatal(err)
	}
	//fmt.Println(episodesFiltered)
	for k, v := range episodesFiltered {
		fmt.Println("key: ", k, " value: ", v)

		for _, m := range v {
			fmt.Print(m)
		}
	}

	jsonString, err := json.Marshal(episodesFiltered)
	//fmt.Println(err)
	var ret []map[string]string
	json.Unmarshal([]byte(string(jsonString)), &ret)

	var i int
	data := [][]string{}
	data = append(data, []string{"Job_Title", "Company_Name", "Location", "Start_Date", "Stipend", "Duration", "Last_Date"})
	for i = 0; i < len(ret); i++ {
		data = append(data, []string{ret[i]["Job_Title"], ret[i]["Company_Name"], ret[i]["Location"], ret[i]["Start_Date"], ret[i]["Stipend"], ret[i]["Duration"], ret[i]["Last_Date"]})

	}
	csvFile, err := os.Create("new.csv")

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)

	for _, empRow := range data {
		_ = csvwriter.Write(empRow)
	}
	csvwriter.Flush()
	csvFile.Close()

	ctx = context.TODO()
	csvfile, err := os.Open("new.csv")
	if err != nil {
		log.Fatal(err)
	}
	df, err := imports.LoadFromCSV(ctx, csvfile, imports.CSVLoadOptions{})
	fmt.Fprint(w, df.Table())
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/newedit", newedithandler)
	myRouter.HandleFunc("/view", viewhandler)

	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func main() {
	handleRequests()
}
