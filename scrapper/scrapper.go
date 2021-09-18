package scrapper

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	ccsv "github.com/tsak/concurrent-csv-writer"
)


type extractedJob struct {
	id string
	location string
	title string
	salary string
	summary string
}


func Scrape(term string) {
	var jobs []extractedJob
	var baseURL = "https://kr.indeed.com/jobs?q="+term
	c := make(chan []extractedJob)
	totalPageLen := getPages(baseURL)
	for i:=0;i < totalPageLen;i++ {
		go getPage(i, baseURL ,c)
	}

	for i:=0;i<totalPageLen;i++ {
		extractedJobs := <- c
		jobs = append(jobs, extractedJobs...)
	}
	writeJob(jobs)
	fmt.Println("I'm done")
}


//go page 1,2,3,4...
func getPage(page int, baseURL string,mc chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)
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
		go extractJob(card, c)
	})
	
	for i:=0;i<searchCards.Length();i++ {
		job := <- c
		jobs = append(jobs, job)
	}
	
	mc <- jobs
}

//extract a Job(extract id,title,location,salary,summary)
func extractJob(card *goquery.Selection,c chan<- extractedJob){
	id,_ := card.Attr("data-jk")
	title := CleanStrings(card.Find(".jobTitle>span").Text())
	location := CleanStrings(card.Find(".companyLocation").Text())
	salary := CleanStrings(card.Find(".salary-snippet").Text())
	summary := CleanStrings(card.Find(".summary").Text())
	c <- extractedJob{
		id: id,
		title: title,
		location: location,
		salary: salary,
		summary: summary,
	}
}

//get all pages
func getPages(baseURL string) int {
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

func writeJob(jobs []extractedJob){
	csv, err := ccsv.NewCsvWriter("job.csv")
	checkError(err)

	defer csv.Close()
	csv.Write([]string{"Link","Title","Location","Salary","Sumary"})
	for _, job := range jobs {
		go func(job extractedJob) {
			csv.Write([]string{"https://kr.indeed.com/viewjob?jk="+job.id, job.title, job.location, job.salary, job.summary})
		}(job)
	}
	for i:=0;i<len(jobs);i++{
		
	}
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
func CleanStrings(str string) string{
	return strings.Join(strings.Fields(strings.TrimSpace(str))," ")
}