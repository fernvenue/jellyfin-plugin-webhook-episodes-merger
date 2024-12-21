package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"
)

type WebhookRequest struct {
	SeriesId      string `json:"SeriesId"`
	SeriesName    string `json:"SeriesName"`
	SeasonNumber  int    `json:"SeasonNumber"`
	EpisodeNumber int    `json:"EpisodeNumber"`
	EpisodeName   string `json:"EpisodeName"`
}

type QueueKey struct {
	SeriesId     string
	SeasonNumber int
}

type Episode struct {
	EpisodeNumber int
	EpisodeName   string
}

type QueueValue struct {
	SeriesName string
	Episodes   []Episode
}

type Config struct {
	ListenAddress    string
	ListenPort       int
	WaitSecond       int
	TextContent      string
	TextKey          string
	EpisodeFormat    string
	TargetURL        string
	AdditionalParams string
	ContentHeader    string
}

var (
	config Config
	queue  = make(map[QueueKey]QueueValue)
	mu     sync.Mutex
)

func main() {
	flag.StringVar(&config.ListenAddress, "listen-address", "::1", "")
	flag.IntVar(&config.ListenPort, "listen-port", 8520, "")
	flag.IntVar(&config.WaitSecond, "wait-second", 300, "")
	flag.StringVar(&config.TextKey, "text-key", "text", "")
	flag.StringVar(&config.TextContent, "text-content", "📺 <b>Episode update reminder:</b> <b>{{.SeriesName}}</b> <b>Season {{.SeasonNumber}}</b>\n", "")
	flag.StringVar(&config.EpisodeFormat, "episode-format", "\nEpisode {{.EpisodeNumber}} {{.EpisodeName}}", "")
	flag.StringVar(&config.TargetURL, "target-url", "", "")
	flag.StringVar(&config.AdditionalParams, "additional-params", "{}", "")
	flag.StringVar(&config.ContentHeader, "content-header", "text", "")
	flag.Parse()

	if config.TargetURL == "" {
		log.Fatal("Error: target-url is required")
	}

	if err := validateJSON(config.AdditionalParams); err != nil {
		log.Fatalf("Invalid JSON in --additional-params: %v", err)
	}

	http.HandleFunc("/", handleWebhook)
	http.HandleFunc("/200", helloWorld)
	address := fmt.Sprintf("[%s]:%d", config.ListenAddress, config.ListenPort)
	log.Printf("Server started at %s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Received request: %v", req)

	key := QueueKey{SeriesId: req.SeriesId, SeasonNumber: req.SeasonNumber}
	mu.Lock()
	value, exists := queue[key]
	if !exists {
		value = QueueValue{SeriesName: req.SeriesName, Episodes: []Episode{}}
	}
	value.Episodes = append(value.Episodes, Episode{
		EpisodeNumber: req.EpisodeNumber,
		EpisodeName:   req.EpisodeName,
	})
	queue[key] = value
	mu.Unlock()

	if !exists {
		log.Printf("Starting to process queue for SeriesId: %s, SeasonNumber: %d", req.SeriesId, req.SeasonNumber)
		go processQueue(key)
	}
	w.WriteHeader(http.StatusOK)
}

func processQueue(key QueueKey) {
	time.Sleep(time.Duration(config.WaitSecond) * time.Second)

	mu.Lock()
	value := queue[key]
	delete(queue, key)
	mu.Unlock()

	log.Printf("Processing queue for SeriesId: %s, SeasonNumber: %d", key.SeriesId, key.SeasonNumber)

	sort.Slice(value.Episodes, func(i, j int) bool {
		return value.Episodes[i].EpisodeNumber < value.Episodes[j].EpisodeNumber
	})

	text, err := buildText(value.SeriesName, key.SeasonNumber, value.Episodes)
	if err != nil {
		log.Printf("Error building text: %v", err)
		return
	}

	params := map[string]interface{}{}
	if err := json.Unmarshal([]byte(config.AdditionalParams), &params); err != nil {
		log.Printf("Error unmarshalling additional params: %v", err)
		return
	}

	tmpl, err := template.New("additionalParams").Parse(config.AdditionalParams)
	if err != nil {
		log.Printf("Error parsing additional params template: %v", err)
		return
	}

	var paramBuf strings.Builder
	err = tmpl.Execute(&paramBuf, struct {
		SeriesId string
	}{
		SeriesId: key.SeriesId,
	})
	if err != nil {
		log.Printf("Error executing additional params template: %v", err)
		return
	}

	finalParams := paramBuf.String()
	if err := json.Unmarshal([]byte(finalParams), &params); err != nil {
		log.Printf("Error unmarshalling final params: %v", err)
		return
	}

	params[config.TextKey] = text

	body, _ := json.Marshal(params)

	log.Printf("Sending request to target URL: %s", config.TargetURL)
	log.Printf("Request body: %s", string(body))

	resp, err := http.Post(config.TargetURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("Error sending request to target URL: %v", err)
		return
	}
	defer resp.Body.Close()

	respBody := new(bytes.Buffer)
	respBody.ReadFrom(resp.Body)

	log.Printf("Response status: %s", resp.Status)
	log.Printf("Response body: %s", respBody.String())
}

func buildText(seriesName string, seasonNumber int, episodes []Episode) (string, error) {
	textTmpl, err := template.New("text").Parse(config.TextContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse text template: %v", err)
	}

	episodeTmpl, err := template.New("episode").Parse(config.EpisodeFormat)
	if err != nil {
		return "", fmt.Errorf("failed to parse episode template: %v", err)
	}

	textData := struct {
		SeriesName   string
		SeasonNumber int
	}{
		SeriesName:   seriesName,
		SeasonNumber: seasonNumber,
	}

	var textBuf strings.Builder
	if err := textTmpl.Execute(&textBuf, textData); err != nil {
		return "", fmt.Errorf("failed to execute text template: %v", err)
	}

	textWithBr := textBuf.String()

	var episodeTextBuf strings.Builder
	for _, ep := range episodes {
		epData := struct {
			EpisodeNumber int
			EpisodeName   string
		}{
			EpisodeNumber: ep.EpisodeNumber,
			EpisodeName:   ep.EpisodeName,
		}
		if err := episodeTmpl.Execute(&episodeTextBuf, epData); err != nil {
			return "", fmt.Errorf("failed to execute episode template: %v", err)
		}
	}

	finalText := textWithBr + episodeTextBuf.String()

	finalText = strings.ReplaceAll(finalText, "\\n", "\n")

	return finalText, nil
}

func validateJSON(input string) error {
	var js map[string]interface{}
	return json.Unmarshal([]byte(input), &js)
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}
