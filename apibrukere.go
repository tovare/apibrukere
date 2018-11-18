package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"

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
	SumEntrances int
	Failed       error
	Widget       bool
	NStilinger   bool
	UsedReferrer url.URL
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
						{StartDate: "2018-11-01", EndDate: "2018-11-15"},
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
	inputdata := make(map[string]Referrer)
	output := make(map[string]Referrer)
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
		tmp, exists := inputdata[domain]
		if !exists {
			tmp := Referrer{}
			tmp.FullReferers = make([]FullReferrer, 1)
		}
		entry := FullReferrer{}
		u, err := url.ParseRequestURI("http://" + partialurl)
		if err != nil {
			log.Fatal(err)
		}
		tmp.Domain = u.Host
		entry.URL = *u
		entry.Entrances, _ = strconv.Atoi(row.Metrics[0].Values[ENTRANCES])
		entry.UniquePageviews, _ = strconv.Atoi(row.Metrics[0].Values[UNIQUEPAGEVIEWS])
		tmp.FullReferers = append(tmp.FullReferers, entry)
		entrances := 0

		for _, ref := range tmp.FullReferers {
			entrances += ref.Entrances
		}
		tmp.SumEntrances = entrances
		// Noen sider har mange forskjellige lenker, vi undersøker den første.
		tmp.UsedReferrer = tmp.FullReferers[0].URL
		inputdata[domain] = tmp

	}

	bar := progressbar.New(len(inputdata))

	client := &http.Client{}
	botget := func(u url.URL, screenshot bool, save bool) (string, error) {
		cmd := "/render/"
		if screenshot {
			cmd = "/screenshot/"
		}
		req, err := http.NewRequest("GET", rendertron+cmd+u.String(), nil)
		log.Println("Kontakter: " + req.URL.String())
		if err != nil {
			log.Println(err)
			return "", err
		}
		req.Header.Set("User-Agent", "NAV tov.are.jacobsen@nav.no")
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return "", err
		}
		if resp.StatusCode != 200 {
			log.Println("StatusCode " + strconv.Itoa(resp.StatusCode) + "for " + u.String())
			return "", errors.New(resp.Status)
		}
		defer resp.Body.Close()

		var page string
		if save {
			if screenshot {
				filename := "out/" + strings.Replace(u.Hostname(), ".", "_", 0) + ".jpg"
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
				return "", err
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return "", err
			}
			page = string(body)
			filename := "out/" + strings.Replace(u.Hostname(), ".", "_", 0) + ".html"
			log.Println("Lagrer " + filename)
			file, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			file.WriteString(page)

		}
		return page, nil
	}

	process := func(wg *sync.WaitGroup, c <-chan Referrer, out chan<- Referrer) {
		for v := range c {
			bar.Add(1)
			resultatstreng, err := botget(v.UsedReferrer, false, true)
			if err != nil {
				v.Failed = err
				out <- v
				continue
			}
			if strings.Contains(resultatstreng, "stillinger/widget") {
				v.Widget = true
				log.Println("Gjorde et funn!")
				botget(v.FullReferers[0].URL, true, true)
			}
			if strings.Count(resultatstreng, "tjenester.nav.no/stillinger") > 1 {
				v.NStilinger = true
				log.Println("Stillinger nevnt mer enn en gang.")
				botget(v.UsedReferrer, true, true)
			}
			out <- v
		}
		wg.Done()
	}

	var bc = make(chan Referrer, 80)
	var br = make(chan Referrer, 10)
	var wg sync.WaitGroup
	var cg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go process(&wg, bc, br)
	}
	consume := func() {
		for v := range br {
			output[v.Domain] = v
		}
		cg.Done()
	}
	cg.Add(1)
	go consume()

	// Fyll køen meed arbeid.
	for _, v := range inputdata {
		bc <- v
	}
	close(bc)
	wg.Wait()
	close(br)
	cg.Wait()
	report(output)

}

func report(resultat map[string]Referrer) {

	type Sum struct {
		AntallDomener   int
		AntallFeil      int
		FlereStillinger int
		AntallWidgets   int
		Detaljer        map[string]Referrer
	}

	t := Sum{}
	t.AntallDomener = len(resultat)
	t.Detaljer = resultat

	for _, v := range resultat {
		if v.Failed != nil {
			t.AntallFeil++
		}
		if v.NStilinger {
			t.FlereStillinger++
		}
		if v.Widget {
			t.AntallWidgets++
		}
	}

	ttext := `

RAPPORT ETTER UNDERSØKELSE
==========================================================================

  Antall domener som er undersøkt......................{{.AntallDomener}}
  Antall domener som ga feilmelding....................{{.AntallFeil}}
  Antall med flere lenker til stillinger...............{{.FlereStillinger}}
  Antall deteksjoner av Widget.........................{{.AntallWidgets}}

BRUKTE WIDGET
==========================================================================

{{range $k, $v := .Detaljer}}{{if $v.Widget}}
* {{$v.Domain}} {{$v.Domain}}{{$v.UsedReferrer.Path}}  ({{$v.SumEntrances}}) {{- end}}{{end}}


DET VAR FLERE LENKER TIL GAMMELT STILLINGSSØK
==========================================================================

{{range $k, $v := .Detaljer}}{{if $v.NStilinger}}
* {{$v.Domain}}{{$v.UsedReferrer.Path}}  ({{$v.SumEntrances}}) {{- end}}{{end}}


DOMENER SOM GA FEIL
==========================================================================

{{range $k, $v := .Detaljer}}{{if $v.Failed}}
* {{$v.Domain}}{{$v.UsedReferrer.Path}}{{- end}}{{end}}
`
	file, err := os.Create("report.md")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	templ, _ := template.New("Rapport").Parse(ttext)
	templ.Execute(file, t)

	s, _ := json.MarshalIndent(resultat, "", "  ")
	log.Println(string(s))

}
