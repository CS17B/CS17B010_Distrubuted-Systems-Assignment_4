package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"

	"time"

	"github.com/gocolly/colly"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//var rnd *renderer.Render

type Fact struct {
	Job_Title    string `json:"Job_Title"`
	Company_Name string `json:"Company_Name"`
	Location     string `json:"Location"`
	Start_Date   string `json:"Start_Date"`
	Duration     string `json:"Duration"`
	Stipend      string `json:"Stipend"`
	Last_Date    string `json:"Last_Date"`
}

func main() {
	allFacts := make([]Fact, 0)

	collector := colly.NewCollector(
		colly.AllowedDomains("internshala.com"),
	)

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://teja1510:teja%401510@cluster0.j2grx.mongodb.net/admin"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 200*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	quickstartDatabase := client.Database("Internship_Details")
	episodesCollection := quickstartDatabase.Collection("Info")
	if err = episodesCollection.Drop(ctx); err != nil {
		log.Fatal(err)
	}

	collector.OnHTML(" .internship_meta", func(e *colly.HTMLElement) {

		facttitle := e.ChildText("div.heading_4_5.profile")
		factcomp := e.ChildText("a.link_display_like_text")
		factloc := e.ChildText("#location_names > span")
		factstar := e.ChildText("span.start_immediately_desktop")
		factduar := e.ChildText("div.item_body")
		factstip := e.ChildText("span.stipend")
		factlast := e.ChildText("div.other_detail_item.apply_by")

		re := regexp.MustCompile(`[0-9][\ ]`)
		fe := regexp.MustCompile(`[0-9]*[\ ][A-Za-z]*['][\ ][0-9]*`)
		newfactduar := re.FindString(factduar) + "Months"
		newfactlast := fe.FindString(factlast)

		if err != nil {
			fmt.Print(err)
		}
		//	fmt.Printf(result)
		fact := Fact{
			Job_Title:    facttitle,
			Company_Name: factcomp,
			Location:     factloc,
			Start_Date:   factstar,
			Duration:     newfactduar,
			Stipend:      factstip,
			Last_Date:    newfactlast,
		}
		quickstartDatabase := client.Database("Internship_Details")
		episodesCollection := quickstartDatabase.Collection("Info")
		var exampleBytes []byte
		exampleBytes, err = json.Marshal(fact)
		var raw map[string]interface{}
		if err := json.Unmarshal(exampleBytes, &raw); err != nil {
			panic(err)
		}
		var going bool = false
		if going == false {
			filterCursor, err := episodesCollection.Find(ctx, bson.M{"Job_Title": facttitle,
				"Company_Name": factcomp,
				"Location":     factloc,
				"Start_Date":   factstar,
				"Duration":     newfactduar,
				"Stipend":      factstip,
				"Last_Date":    newfactlast})
			if err != nil {
				log.Fatal(err)
			}
			var x int = 0
			for filterCursor.Next(ctx) {
				x++
			}
			if x == 0 {
				_, err := episodesCollection.InsertOne(ctx, raw)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}

		allFacts = append(allFacts, fact)
	})
	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting", request.URL.String())
	})
	for i := 0; i < 2; i++ {
		fmt.Printf("Scraping Page : %d\n", i)

		collector.Visit("https://internshala.com/internships/internship-in-bangalore/page-" + strconv.Itoa(i))

		log.Printf("Scrapping Complete\n")

	}
	writeJSON(allFacts)

	//time.Sleep(100 * time.Second)

}

func writeJSON(data []Fact) {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Println("Unable to create json file")
		return
	}
	_ = ioutil.WriteFile("Data.json", file, 0644)
}
