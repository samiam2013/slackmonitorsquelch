package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	logger "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

func squelcher(thresholds networkThresholds) {
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
	if signingSecret == "" {
		logger.Fatal("SLACK_SIGNING_SECRET environment variable not set")
	}

	http.HandleFunc("/squelch", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received a request")

		verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = verifier.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var response string
		switch s.Command {
		case "/squelch":
			tokens := strings.Fields(s.Text)
			if len(tokens) < 2 {
				response = "Usage: /squelch <network ID> <threshold in seconds>"
			}
			networkID, err := strconv.Atoi(tokens[0])
			if err != nil {
				response = "Network ID must be an integer"
			}
			threshold, err := strconv.Atoi(tokens[1])
			if err != nil {
				response = "Threshold must be an integer"
			}
			thresholds.Lock()
			thresholds.thresholds[int64(networkID)] = int64(threshold)
			thresholds.Unlock()
			response = fmt.Sprintf("Set squelch threshold for network %d to %d seconds",
				networkID, threshold)
		default:
			response = "Unknown command"
		}
		params := &slack.Msg{Text: response}
		b, err := json.Marshal(params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)

	})
	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":3000", nil)
}
