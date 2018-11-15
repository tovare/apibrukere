package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"
	"strings"

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
	)
	flag.StringVar(&credentials, "c", "private/tovare-a7a5db068b79.json", "Google API credentials")
	flag.IntVar(&antallResultater, "n", 10, "Antall restulater å analysere, max 100 000")
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

	// Crawl alle sider. Noen sider har mange forskjellige lenker.
	// Dette løses ved å plukke en tilfeldig lenke.

	//	for k, v := range resultat {
	//
	//	}

	s, _ := json.MarshalIndent(resultat, "", "  ")
	log.Println(string(s))

}
