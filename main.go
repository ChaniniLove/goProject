package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var baseURL = "https://kr.indeed.com/jobs?q=golang"

type extractedJob struct {
	id string
	location string
	title string
	salary string
	summary string
}


func main() {
	var jobs []extractedJob
	totalLen := getPages()
	for i:=0;i < totalLen;i++ {
		extractedJobs := getPage(i)
		jobs = append(jobs, extractedJobs...)
	}
	writeJob(jobs)
}

func writeJob(jobs []extractedJob){
	file, err := os.Create("jobs.csv")
	checkError(err)
	utf8bom := []byte{0xEF, 0xBB, 0xBF}
	file.Write(utf8bom)

	w := csv.NewWriter(file)
	defer w.Flush()
	
	header := []string{"ID","TITLE","LOCATION","SALARY","SUMMARY"}
	wErr := w.Write(header)
	checkError(wErr)

	for _, job := range jobs {
		jobSlice := []string{job.id, job.title, job.location, job.salary, job.summary}
		jse := w.Write(jobSlice)
		checkError(jse)
	}
}

//go page 1,2,3,4...
func getPage(page int) []extractedJob{
	var jobs []extractedJob
	pageURL := baseURL + "&start=" + strconv.Itoa(page*50)
	fmt.Println("Requesting:",pageURL)

	res, err := http.Get(pageURL)
	checkError(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	searchCards := doc.Find(".tapItem")
	searchCards.Each(func(_ int, card *goquery.Selection){
		job := extractJob(card)
		jobs = append(jobs, job)
	})
	return jobs
}

//extract a Job(extract id,title,location,salary,summary)
func extractJob(card *goquery.Selection) extractedJob{
	id,_ := card.Attr("data-jk")
	title := cleanStrings(card.Find(".jobTitle>span").Text())
	location := cleanStrings(card.Find(".companyLocation").Text())
	salary := cleanStrings(card.Find(".salary-snippet").Text())
	summary := cleanStrings(card.Find(".summary").Text())
	return extractedJob{
		id: id,
		title: title,
		location: location,
		salary: salary,
		summary: summary,
	}
}

//get all pages
func getPages() int {
	pages := 0
	res, err := http.Get(baseURL)
	checkError(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)
	doc.Find(".pagination").Each(func(_ int, s *goquery.Selection){
		pages = s.Find("a").Length()
	})
	return pages
}

//check error
func checkError(err error){
	if err != nil{
		log.Fatal(err)
	}
}

//check status code
func checkCode(res *http.Response){
	if res.StatusCode != 200{
		log.Fatal("Status code isnt 200:",res.StatusCode)
	}
}

//make clean string
func cleanStrings(str string) string{
	return strings.Join(strings.Fields(strings.TrimSpace(str))," ")
}