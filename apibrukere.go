package main

import (
	"context"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/schollz/progressbar"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	ga "google.golang.org/api/analyticsreporting/v4"
)

// FullReferrer ... Hvert resultat
type FullReferrer struct {
	URL             url.URL
	Entrances       int
	UniquePageviews int
}

// Referrer ... Hvert domene
type Referrer struct {
	Domain       string
	FullReferers []FullReferrer
}

func main() {
	var (
		credentials      string
		antallResultater int
		saveall          bool
		rendertron       string
	)
	flag.StringVar(&credentials, "c", "private/tovare-a7a5db068b79.json", "Google API credentials")
	flag.IntVar(&antallResultater, "n", 10, "Antall restulater å analysere, max 100 000")
	flag.BoolVar(&saveall, "s", false, "Lagrer tekst og bilde fra alle sider")
	flag.StringVar(&rendertron, "r", "http://127.0.0.1:3000", "Adressen til Google rendertron")
	flag.Parse()

	// Autentiser Analytics Reporting API.
	var service *ga.Service

	{
		ctx := context.Background()
		data, err := ioutil.ReadFile(credentials)
		if err != nil {
			log.Fatal(err)
		}
		creds, err := google.CredentialsFromJSON(ctx, data, ga.AnalyticsReadonlyScope)
		if err != nil {
			log.Fatal(err)
		}
		service, err = ga.New(oauth2.NewClient(ctx, creds.TokenSource))
		if err != nil {
			log.Fatal(err)
		}
	}

	// Definer og kjør rapport
	var myreport *ga.GetReportsResponse

	{
		req := &ga.GetReportsRequest{
			ReportRequests: []*ga.ReportRequest{
				{
					ViewId: "95725034",

					DateRanges: []*ga.DateRange{
						{StartDate: "2018-08-01", EndDate: "2018-08-02"},
					},
					Metrics: []*ga.Metric{
						{Expression: "ga:entrances"},
						{Expression: "ga:uniquePageviews"},
					},
					Dimensions: []*ga.Dimension{
						{Name: "ga:fullReferrer"},
					},
					PageSize: int64(antallResultater),
				},
			},
		}
		var err error
		myreport, err = service.Reports.BatchGet(req).Do()
		if err != nil {
			log.Fatal(err)
		}
	}

	// Lag en liste over alle henvisninger i resultat
	resultat := make(map[string]Referrer)
	const (
		ENTRANCES = iota
		UNIQUEPAGEVIEWS
	)

	for _, row := range myreport.Reports[0].Data.Rows {
		// Eksepler: 69nord.no/ledige-stillinger , (direct)
		partialurl := row.Dimensions[0]
		domain := strings.Split(partialurl, "/")[0]
		if domain == "(direct)" {
			continue
		}
		tmp, exists := resultat[domain]
		if !exists {
			tmp := Referrer{}
			tmp.FullReferers = make([]FullReferrer, 1)
		}
		entry := FullReferrer{}
		u, err := url.ParseRequestURI("http://" + partialurl)
		if err != nil {
			log.Fatal(err)
		}
		entry.URL = *u
		entry.Entrances, _ = strconv.Atoi(row.Metrics[0].Values[ENTRANCES])
		entry.UniquePageviews, _ = strconv.Atoi(row.Metrics[0].Values[UNIQUEPAGEVIEWS])
		tmp.FullReferers = append(tmp.FullReferers, entry)
		resultat[domain] = tmp
	}

	bar := progressbar.NewOptions(len(resultat),
		progressbar.OptionSetWriter(os.Stderr))

	// Crawl alle sider. Noen sider har mange forskjellige lenker.
	// Dette løses ved å plukke en tilfeldig lenke.

	client := &http.Client{}
	botget := func(u url.URL, screenshot bool, save bool) string {
		cmd := "/render/"
		if screenshot {
			cmd = "/screenshot/"
		}
		req, err := http.NewRequest("GET", rendertron+cmd+u.String(), nil)
		log.Println("Kontakter: " + req.URL.String())
		if err != nil {
			log.Println(err)
			return ""
		}
		req.Header.Set("User-Agent", "NAV tov.are.jacobsen@nav.no")
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return ""
		}
		if resp.StatusCode != 200 {
			log.Println("StatusCode " + strconv.Itoa(resp.StatusCode) + "for " + u.String())
			return ""
		}
		defer resp.Body.Close()
		var page string
		if save {
			if screenshot {
				filename := strings.Replace(u.Hostname(), ".", "_", 0) + ".jpg"
				log.Println("Lagrer " + filename)
				file, err := os.Create(filename)
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()
				_, err = io.Copy(file, resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				return "image"
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return ""
			}
			page = string(body)
			filename := strings.Replace(u.Hostname(), ".", "_", 0) + ".html"
			log.Println("Lagrer " + filename)
			file, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			file.WriteString(page)

		}
		return page
	}

	for _, v := range resultat {
		resultatstreng := botget(v.FullReferers[0].URL, true, true)
		if strings.Contains(resultatstreng, "stillinger/widget") {
			log.Println("Gjorde et funn!")
			botget(v.FullReferers[0].URL, true, true)
		}
		bar.Add(1)
	}

	//s, _ := json.MarshalIndent(resultat, "", "  ")
	//log.Println(string(s))

}
