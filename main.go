package main

import (
	"os"
	"strings"

	"github.com/ChaniniLove/myGoProject/scrapper"
	"github.com/labstack/echo"
)

func handleHome(c echo.Context) error {
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	term := strings.ToLower(scrapper.CleanStrings(c.FormValue("term")))
	scrapper.Scrape(term)
	defer os.Remove("job.csv")
	return c.Attachment("job.csv","job.csv")
}

func main() {
	e := echo.New()
	e.GET("/",handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))
}