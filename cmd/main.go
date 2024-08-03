package main

import (
	"client"
	"context"
	"log"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/joho/godotenv"
)

func main() {
	// Construct a new API object using a global API key
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	api, err := cloudflare.New(os.Getenv("CLOUDFLARE_API_KEY"), os.Getenv("CLOUDFLARE_API_EMAIL"))
	// alternatively, you can use a scoped API token
	// api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	// Most API calls require a Context
	ctx := context.Background()

	// Fetch user details on the account
	rc := cloudflare.AccountIdentifier(os.Getenv("CLOUDFLARE_ACCOUNT_ID"))
	tID := os.Getenv("CLOUDFLARE_TUNNEL_ID")
	zID := os.Getenv("CLOUDFLARE_ZONE_ID")
	tunnelConfig, err := api.GetTunnelConfiguration(ctx, rc, tID)
	if err != nil {
		log.Fatal(err)
	}
	// Print user details
	var data []cloudflare.UnvalidatedIngressRule
	for _, item := range tunnelConfig.Config.Ingress {
		if strings.HasPrefix(item.Service, "rdp") {
			data = append(data, item)
		}
	}

	client.StartApp(data, api, ctx, rc, tID, zID)
}
