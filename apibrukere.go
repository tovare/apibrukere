package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2"

	"golang.org/x/oauth2/google"
	ga "google.golang.org/api/analyticsreporting/v4"
)

func main() {
	var (
		credentials      string
		antallResultater int
	)
	flag.StringVar(&credentials, "c", "private/tovare-a7a5db068b79.json", "Google API credentials")
	flag.IntVar(&antallResultater, "n", 10, "Antall restulater å analysere, max 100 000")
	flag.Parse()

	// Autentiser Analytics Reporting API
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

	log.Println("Completed")
	log.Println(myreport.Reports[0].Data.Rows[1])

}
