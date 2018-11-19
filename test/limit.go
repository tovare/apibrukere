package main

import (
	"context"
	"log"

	"golang.org/x/time/rate"
)

func main() {
	limit := rate.NewLimiter(1, 1)
	ctx := context.Background()
	for true {
		limit.Wait(ctx)
		log.Println("Hello")
	}

}
