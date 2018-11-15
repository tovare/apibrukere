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
		credentials string
	)
	flag.StringVar(&credentials, "c", "private/tovare-a7a5db068b79.json", "Google API credentials")
	flag.Parse()

	// Autentiser Analytics Reporting API
	ctx := context.Background()
	data, err := ioutil.ReadFile(credentials)
	if err != nil {
		log.Fatal(err)
	}
	creds, err := google.CredentialsFromJSON(ctx, data, ga.AnalyticsReadonlyScope)
	if err != nil {
		log.Fatal(err)
	}
	service, err := ga.New(oauth2.NewClient(ctx, creds.TokenSource))
	if err != nil {
		log.Fatal(err)
	}

	// Kjør rapport
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
				PageSize: 10,
			},
		},
	}

	myreport, err := service.Reports.BatchGet(req).Do()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Completed")
	log.Println(myreport.Reports[0].Data.Rows[1])

}
