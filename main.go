package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.UserDataDir("someUserDir"),
		chromedp.Flag("headless", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("restore-on-startup", false),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)

	ctx, _ := chromedp.NewContext(allocCtx)
	if err := chromedp.Run(ctx, chromedp.Navigate("about:blank")); err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/bankid", func(w http.ResponseWriter, r *http.Request) {
		c := make(chan string, 1)

		listenForNetworkEvent(ctx, c)
		navigateToBankID(ctx)

		tokenValue := <-c
		http.Redirect(w, r, "bankid:///?autostarttoken="+tokenValue+"&redirect=null", http.StatusSeeOther)
	})

	log.Printf("Starting server at port 8000\n")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}

func navigateToBankID(ctx context.Context) {
	err := chromedp.Run(ctx,
		chromedp.Navigate(`<URL of BankID page>`),
		// chromedp.Click(`#login-button`, chromedp.NodeVisible),
	)
	if err != nil {
		log.Fatal(err)
	}
}

// If you need to catch network events, i.e. javascript calls, if the autostarttoken is not available in the DOM
func listenForNetworkEvent(ctx context.Context, channel chan string) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventResponseReceived:
			// ev.Type is usually "XHR" or "Fetch"
			if ev.Type == "XHR" {
				resp := ev.Response
				if resp.URL == "<URL to filter on which contains the autostarttoken>" {
					go func() {
						c := chromedp.FromContext(ctx)
						rbp := network.GetResponseBody(ev.RequestID)
						body, err := rbp.Do(cdp.WithExecutor(ctx, c.Target))
						if err != nil {
							log.Fatal(err)
						}
						if err = os.WriteFile(ev.RequestID.String(), body, 0644); err != nil {
							log.Fatal(err)
						}
						// Parse the response body to get the autostarttoken, sometimes it's enough with a map other times it's easier to create a struct
						jsonMap := make(map[string]interface{})
						json.Unmarshal(body, &jsonMap)
						token := jsonMap["autoStartToken"].(string)

						channel <- token
					}()
				}
			}
		}
	})
}

