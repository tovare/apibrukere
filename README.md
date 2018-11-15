# apibrukere
Analyse av hvem som bruker widget fra stillingssoket


## Mål: Kartlegge flest mulig av windget- og API-brukerne før nyttår.

## Planen

1. Hent inn liste over referrers med volum fra Google Analytics.
2. Besøk kilden og sjekk om de bruker widget.
3. Lagre skjermbildet

# Tekniske detaljer rendering

Prosjektet bruker rendertron som startes med:

     PORT=8080 npm run start

Nb: puppeteer må installeres seperat. Config til typescript lint måtte endres til å bli mindre aggressiv.

## Bildegalleri

 montage  -thumbnail 300x300  -bordercolor grey  -background grey40 -pointsize 9 -density 144x144 +polaroid -resize 50%  -background white -geometry +1+1 -tile 10x10 -set caption %t -title "Sider som lenker til stillinger" *.jpg polaroid_t.jpg



# Tekniske detaljer Gogle Analytics

For å sette opp må det lages et prosjekt og lage json-credentials til Google Analytics API v4 som legges i en private/ katalog (ikke på github). Tilgangen trenger leserettigheter.

https://console.cloud.google.com/apis/api/analyticsreporting.googleapis.com/overview?project=tovare-222514

Account legges på viewer:

tovarecrawler@tovare-222514.iam.gserviceaccount.com har view tilgang på 02 Stillingssok

https://developers.google.com/analytics/devguides/reporting/core/dimsmets#cats=traffic_sources,page_tracking


## Google Analytics API request

POST https://analyticsreporting.googleapis.com/v4/reports:batchGet?key={YOUR_API_KEY}

    {
    "reportRequests": [
    {
    "dateRanges": [
        {
        "startDate": "2018-08-01",
        "endDate": "2018-08-02"
        }
    ],
    "hideTotals": true,
    "viewId": "95725034",
    "dimensions": [
        {
        "name": "ga:fullReferrer"
        }
    ],
    "metrics": [
        {
        "expression": "ga:entrances"
        },
        {
        "expression": "ga:uniquePageviews"
        }
    ],
    "pageSize": 10
    }
    ]



### RESPONSE

    cache-control:  private
    content-encoding:  gzip
    content-length:  662
    content-type:  application/json; charset=UTF-8
    date:  Wed, 14 Nov 2018 15:36:46 GMT
    server:  ESF
    vary:  Origin, X-Origin, Referer
    
    {
    "reports": [
    {
    "columnHeader": {
        "dimensions": [
        "ga:fullReferrer"
        ],
        "metricHeader": {
        "metricHeaderEntries": [
        {
        "name": "ga:entrances",
        "type": "INTEGER"
        },
        {
        "name": "ga:uniquePageviews",
        "type": "INTEGER"
        }
        ]
        }
    },
    "data": {
        "rows": [
        {
        "dimensions": [
        "(direct)"
        ],
        "metrics": [
        {
            "values": [
            "27768",
            "90710"
            ]
        }
        ]
        },
        {
        "dimensions": [
        "1881.no/nav/nav-hordaland/nav-norheimsund/nav-kvam_101336032S1/"
        ],
        "metrics": [
        {
            "values": [
            "2",
            "5"
            ]
        }
        ]
        },
        {
        "dimensions": [
        "1881.no/nav/nav-moere-og-romsdal/nav-oersta/nav-oersta-volda_101336781S1/"
        ],
        "metrics": [
        {
            "values": [
            "2",
            "7"
            ]
        }
        ]
        },
        {
        "dimensions": [
        "1881.no/nav/nav-nordland/nav-bodoe/nav-bodoe_103143373S1/"
        ],
        "metrics": [
        {
            "values": [
            "1",
            "3"
            ]
        }
        ]
        },
        {
        "dimensions": [
        "1881.no/nav/nav-nordland/nav-bodoe/nav-hjelpemiddelsentral-nordland_100450579S1/"
        ],
        "metrics": [
        {
            "values": [
            "1",
            "7"
            ]
        }
        ]
        },
        {
        "dimensions": [
        "1881.no/nav/nav-oslo/nav-oslo-sentrum/nav-innkreving_100106888S20/"
        ],
        "metrics": [
        {
            "values": [
            "1",
            "9"
            ]
        }
        ]
        },
        {
        "dimensions": [
        "1881.no/nav/nav-sogn-og-fjordane/nav-sogndal/nav-sogndal_101337003S1/"
        ],
        "metrics": [
        {
            "values": [
            "1",
            "1"
            ]
        }
        ]
        },
        {
        "dimensions": [
        "1881.no/nav/nav-troms/nav-harstad/nav-harstad_100121466S1/"
        ],
        "metrics": [
        {
            "values": [
            "1",
            "2"
            ]
        }
        ]
        },
        {
        "dimensions": [
        "1881.no/sosiale-tjenester-sosialkontor-og-sentre/sosiale-tjenester-sosialkontor-og-sentre-troendelag/sosiale-tjenester-sosialkontor-og-sentre-trondheim/nav-midtbyen_101408420S1/"
        ],
        "metrics": [
        {
            "values": [
            "3",
            "16"
            ]
        }
        ]
        },
        {
        "dimensions": [
        "69nord.no/ledige-stillinger"
        ],
        "metrics": [
        {
            "values": [
            "1",
            "1"
            ]
        }
        ]
        }
        ],
        "rowCount": 1789,
        "minimums": [
        {
        "values": [
        "0",
        "0"
        ]
        }
        ],
        "maximums": [
        {
        "values": [
        "27768",
        "100583"
        ]
        }
        ],
        "isDataGolden": true
    },
    "nextPageToken": "10"
    }
    ]
    }

