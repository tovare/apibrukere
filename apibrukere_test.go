package main

import (
	"context"
	"io/ioutil"
	"log"
	"testing"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	ga "google.golang.org/api/analyticsreporting/v4"
)

// For å teste tidsforbruk og sampling av GA-resultater.
func BenchmarkGoogleAPI(b *testing.B) {
	// Autentiser
	var service *ga.Service
	ctx := context.Background()
	data, err := ioutil.ReadFile("private/tovare-a7a5db068b79.json")
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

	// Definer og kjør rapport
	var myreport *ga.GetReportsResponse
	req := &ga.GetReportsRequest{
		ReportRequests: []*ga.ReportRequest{
			{
				ViewId: "95725034",

				DateRanges: []*ga.DateRange{
					{StartDate: "2018-07-01", EndDate: "2018-07-02"},
				},
				Metrics: []*ga.Metric{
					{Expression: "ga:entrances"},
					{Expression: "ga:uniquePageviews"},
				},
				Dimensions: []*ga.Dimension{
					{Name: "ga:fullReferrer"},
				},
				PageSize: int64(2000),
			},
		},
	}
	myreport, err = service.Reports.BatchGet(req).Do()
	if err != nil {
		log.Fatal(err)
	}
	// Viser at jeg får sampling hvis jeg tar mer enn 2 dager.
	log.Println(len(myreport.Reports[0].Data.Rows))
	log.Printf("Next page %s", myreport.Reports[0].NextPageToken)
	log.Printf("Sampling ReadCounts %v, SpaceSizes %v",
		myreport.Reports[0].Data.SamplesReadCounts,
		myreport.Reports[0].Data.SamplingSpaceSizes)
}
