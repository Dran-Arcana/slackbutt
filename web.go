package main

// Slack outgoing webhooks are handled here. Requests come in and are run through
// the markov chain to generate a response, which is sent back to Slack.
//
// Create an outgoing webhook in your Slack here:
// https://my.slack.com/services/new/outgoing-webhook

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	"regexp"
)

type WebhookResponse struct {
	Username string `json:"username"`
	Text     string `json:"text"`
}

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		incomingText := r.PostFormValue("text")
		if incomingText != "" && r.PostFormValue("user_id") != "" {
			text := parseText(incomingText)
			log.Printf("Handling incoming request: %s", text)

			if text != "" {
				markovChain.Write(text)
			}

			go func() {
				markovChain.Save(stateFile)
			}()

			if rand.Intn(100) <= responseChance || strings.Contains(text, botUsername) {
				var startStr string
				var matchStr string
				startStr = ""
				// https://regex-golang.appspot.com/assets/html/index.html
				wiy := regexp.MustCompile("([W|w]hat is your)([0-9A-Za-z_]*( )*)*")
				if wiy.FindString(text) !="" {
					matchStr = strings.Split(wiy.FindString(text), "hat is your")[1]
					startStr = "My "
					append(startStr, matchStr)
				}
				var response WebhookResponse
				response.Username = botUsername
				response.Text = markovChain.Generate(numWords, startStr)
				log.Printf("Sending response: %s", response.Text)

				b, err := json.Marshal(response)
				if err != nil {
					log.Fatal(err)
				}

				if twitterClient != nil {
					log.Printf("Tweeting: %s", response.Text)
					twitterClient.Post(response.Text)
				}

				time.Sleep(5 * time.Second)
				w.Write(b)
			}
		}
	})
}

func StartServer(port int) {
	log.Printf("Starting HTTP server on %d", port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
