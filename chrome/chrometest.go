//
// Test if chromedp is a viable alternative
//
package main

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/chromedp/chromedp"
)

func main() {
	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// run task list
	var buf []byte
	err := chromedp.Run(ctx, screenshot(`https://tovare.com`, &buf))
	if err != nil {
		log.Fatal(err)
	}

	// save the screenshot to disk
	if err = ioutil.WriteFile("screenshot.png", buf, 0644); err != nil {
		log.Fatal(err)
	}
}

func screenshot(urlstr string, res *[]byte) chromedp.Tasks {
	log.Println("ok")
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Screenshot(nil, res),
	}
}
