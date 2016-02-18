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

var lastResponse string

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		incomingText := r.PostFormValue("text")
		if incomingText != "" && r.PostFormValue("user_id") != "" {
			text := parseText(incomingText)
			log.Printf("Handling incoming request: %s", text)

			if rand.Intn(99) < responseChance || strings.Contains(text, botUsername) {
				var startStr string
				var matchStr string
				var response WebhookResponse
				response.Username = botUsername
				markovBrute := 1
				markovBruteFound := false
				startStr = ""
				
				// https://regex-golang.appspot.com/assets/html/index.html
				
				if strings.Contains(text, "What is your") || strings.Contains(text, "what is your") {
					wiy := regexp.MustCompile("([W|w]hat is your)([0-9A-Za-z_]*( )*)*")
					matchStr = strings.Trim(strings.Split(wiy.FindString(text), "hat is your")[1], " ")
					if rand.Intn(99) < 90 {
						startStr = strings.Trim("My " + matchStr +" is", " ")
					} else {
						startStr = strings.Trim("My " + matchStr, " ")
					}
					log.Printf("  \\----Handling special request: what is your |")
					log.Printf("    \\----Handling special request: smart |")
					log.Printf("    \\----matchStr:|%s|", matchStr)
					log.Printf("    \\----startStr:|%s|", startStr)
					
					markovBruteFound = false
					
					strSplit := strings.Split(startStr, " ")
					backupBrute := strSplit[len(strSplit)-1]
					for markovBrute < 25000 && markovBruteFound == false {
						markovBrute += 1
						response.Text = markovChain.Generate(numWords, "")
						log.Printf("      \\----trying:|%i|%s|", markovBrute, response.Text)
						if strings.Contains(response.Text, startStr) || strings.Contains(response.Text, matchStr){
							log.Printf("        \\----found!:|%s|", response.Text)
							markovBruteFound = true
							break
						}
						if strings.Contains(response.Text, backupBrute){
							backupBrute = response.Text	
						}
					}
					if backupBrute != strSplit[len(strSplit)-1]{
						response.Text = backupBrute	
					}
					
				} else {
					response.Text = markovChain.Generate(numWords, "")	
				}
				
				log.Printf("Sending response: %s", response.Text)

				b, err := json.Marshal(response)
				if err != nil {
					log.Fatal(err)
				}

				if twitterClient != nil {
					log.Printf("Tweeting: %s", response.Text)
					twitterClient.Post(response.Text)
				}

				time.Sleep(1 * time.Second)
				w.Write(b)
				lastResponse = response.Text
				
			} else {
				if text != "" && !strings.Contains(lastResponse, text){
					log.Printf("  \\----learning: yes")
					markovChain.Write(text)
				}

				go func() {
					markovChain.Save(stateFile)
				}()
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
