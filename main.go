package main

import (
	"fmt"
	log "log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	mf "miniflux.app/v2/client"
)

type minifluxClient interface {
	Entries(*mf.Filter) (*mf.EntryResultSet, error)
	UpdateEntries([]int64, string) error
}

type httpClient interface {
	Head(string) (*http.Response, error)
}

func main() {
	mfClient := mf.NewClient(os.Getenv("MINIFLUX_ENDPOINT"), os.Getenv("MINIFLUX_TOKEN"))

	// REMOVE_SHORTS defaults to true
	removeShortsEnv, present := os.LookupEnv("REMOVE_SHORTS")
	if !present {
		removeShortsEnv = "1"
	}
	removeShorts, err := strconv.ParseBool(removeShortsEnv)
	if err != nil {
		panic("REMOVE_SHORTS could not be parsed as a boolean")
	}

	// REMOVE_LIVES defaults to true
	removeLivesEnv, present := os.LookupEnv("REMOVE_LIVES")
	if !present {
		removeLivesEnv = "1"
	}
	removeLives, err := strconv.ParseBool(removeLivesEnv)
	if err != nil {
		panic("REMOVE_LIVES could not be parsed as a boolean")
	}

	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	clean(mfClient, removeShorts, removeLives, httpClient)
}

func clean(mfClient minifluxClient, removeShorts bool, removeLives bool, httpClient httpClient) {
	if !(removeShorts || removeLives) {
		log.Warn("Nothing needs to be removed, skipping cleanup")
		return
	}

	log.Info("Beginning cleanup")
	entries, _ := mfClient.Entries(&mf.Filter{
		Status: "unread",
	})

	ytEntries := mf.Entries{}
	for _, entry := range entries.Entries {
		if strings.Contains(entry.URL, "youtube.com") {
			ytEntries = append(ytEntries, entry)
		}
	}

	idsToRemove := []int64{}
	for _, entry := range ytEntries {
		isShort := isShort(httpClient, entry)
		if removeShorts && isShort {
			idsToRemove = append(idsToRemove, entry.ID)
		} else if removeLives && isLive(entry) && !isShort {
			idsToRemove = append(idsToRemove, entry.ID)
		}
	}

	if len(idsToRemove) == 0 {
		log.Info("No entries to remove")
	} else {
		if len(idsToRemove) == 1 {
			log.Info("Removing 1 entry")
		} else {
			log.Info(fmt.Sprintf("Removing %d entries", len(idsToRemove)))
		}
		mfClient.UpdateEntries(idsToRemove, "removed")
	}

	log.Info("Cleanup complete")
}

func isShort(client httpClient, entry *mf.Entry) bool {
	// normal videos will redirect from the following URL, shorts will not
	redirects, _ := doesRedirect(client, "https://www.youtube.com/shorts/"+getVideoID(entry.URL))
	return !redirects
}

func getVideoID(url string) string {
	return strings.Split(url, "?v=")[1]
}

func doesRedirect(client httpClient, url string) (bool, error) {
	response, err := client.Head(url)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 && response.StatusCode < 400 {
		return true, nil
	}
	return false, nil
}

func isLive(entry *mf.Entry) bool {
	// miniflux is not able to scrape reading times from live streams
	return entry.ReadingTime == 0
}
