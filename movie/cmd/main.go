package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/mamalmaleki/go_movie/movie/internal/controller/movie"
	metadataGatewayPkg "github.com/mamalmaleki/go_movie/movie/internal/gateway/metadata/http"
	ratingGatewayPkg "github.com/mamalmaleki/go_movie/movie/internal/gateway/rating/http"
	httpHandler "github.com/mamalmaleki/go_movie/movie/internal/handler/http"
	"github.com/mamalmaleki/go_movie/pkg/discovery"
	"github.com/mamalmaleki/go_movie/pkg/discovery/consul"
	"log"
	"net/http"
	"time"
)

const serviceName = "movie"

func main() {
	var port int
	flag.IntVar(&port, "port", 8083, "API handler port")
	flag.Parse()
	log.Printf("Starting the movie metadata service on port %d", port)
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName,
		fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	metadataGateway := metadataGatewayPkg.New(registry)
	ratingGateway := ratingGatewayPkg.New(registry)
	ctrl := movie.New(ratingGateway, metadataGateway)
	h := httpHandler.New(ctrl)
	http.Handle("/movie", http.HandlerFunc(h.GetMovieDetails))
	if err := http.ListenAndServe(":8083", nil); err != nil {
		panic(err)
	}
}
