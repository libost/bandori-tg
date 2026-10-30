package utils

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/libost/bandori-tg/config"
	C "github.com/libost/bandori-tg/constants"
	"golang.org/x/net/proxy"
	"golang.org/x/sync/errgroup"
)

var HttpClient *http.Client

type fetchTask struct {
	URL    string
	Target any
}

var dataMu sync.RWMutex

var Band C.BandData
var Cards C.CardData
var Characters C.CharacterData
var Events C.EventsData
var Recent C.Recent
var Skills C.SkillData
var Gacha C.GachaData
var Rates C.RatesData
var Songs C.SongsData
var Comic C.ComicData
var ComicKeys []string
var REventsKeys []string
var RGachaKeys []string
var SongsKeysJP []string
var SongsKeysEN []string
var SongsKeysCN []string
var SongsKeysKR []string
var SongsKeysTW []string

func ReadBand() C.BandData {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Band
}

func ReadCards() C.CardData {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Cards
}

func ReadCharacters() C.CharacterData {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Characters
}

func ReadEvents() C.EventsData {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Events
}

func ReadRecent() C.Recent {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Recent
}

func ReadSkills() C.SkillData {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Skills
}

func ReadGacha() C.GachaData {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Gacha
}

func ReadRates() C.RatesData {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Rates
}

func ReadSongs() C.SongsData {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Songs
}

func ReadComic() C.ComicData {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Comic
}

func ReadREventsKeys() []string {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return append([]string(nil), REventsKeys...)
}

func ReadRGachasKeys() []string {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return append([]string(nil), RGachaKeys...)
}

func ReadComicKeys() []string {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return append([]string(nil), ComicKeys...)
}

func publishDataSnapshot(snapshotBand C.BandData, snapshotCards C.CardData, snapshotCharacters C.CharacterData, snapshotEvents C.EventsData, snapshotRecent C.Recent, snapshotSkills C.SkillData, snapshotGacha C.GachaData, snapshotRates C.RatesData, snapshotSongs C.SongsData, snapshotComic C.ComicData) {
	dataMu.Lock()
	defer dataMu.Unlock()

	Band = snapshotBand
	Cards = snapshotCards
	Characters = snapshotCharacters
	Events = snapshotEvents
	Recent = snapshotRecent
	Skills = snapshotSkills
	Gacha = snapshotGacha
	Rates = snapshotRates
	Songs = snapshotSongs
	Comic = snapshotComic

	REventsKeys = make([]string, 0, len(Recent.Events))
	for k := range Recent.Events {
		REventsKeys = append(REventsKeys, k)
	}
	sort.Strings(REventsKeys)

	RGachaKeys = make([]string, 0, len(Recent.Gacha))
	for k := range Recent.Gacha {
		RGachaKeys = append(RGachaKeys, k)
	}
	sort.Strings(RGachaKeys)

	SongsKeysJP = make([]string, 0, len(Songs))
	SongsKeysCN = make([]string, 0, len(Songs))
	SongsKeysEN = make([]string, 0, len(Songs))
	SongsKeysKR = make([]string, 0, len(Songs))
	SongsKeysTW = make([]string, 0, len(Songs))
	for k := range Songs {
		SongsKeysJP = append(SongsKeysJP, Songs[k].MusicTitle[0])
		SongsKeysEN = append(SongsKeysEN, Songs[k].MusicTitle[1])
		SongsKeysTW = append(SongsKeysTW, Songs[k].MusicTitle[2])
		SongsKeysCN = append(SongsKeysCN, Songs[k].MusicTitle[3])
		SongsKeysKR = append(SongsKeysKR, Songs[k].MusicTitle[4])
	}
	sort.Strings(SongsKeysJP)

	sort.Strings(SongsKeysEN)

	sort.Strings(SongsKeysCN)

	sort.Strings(SongsKeysKR)

	sort.Strings(SongsKeysTW)

	ComicKeys = make([]string, 0, len(Comic))
	for k := range Comic {
		ComicKeys = append(ComicKeys, k)
	}

	sort.Strings(ComicKeys)
}

func InitLists() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config := config.AppConfig

	if config.Proxy.Enabled {
		switch config.Proxy.Type {
		case "http":
			HttpClient = &http.Client{
				Timeout: 5 * time.Second,
				Transport: &http.Transport{
					MaxIdleConnsPerHost: 10,
					Proxy: http.ProxyURL(&url.URL{
						Scheme: "http",
						Host:   fmt.Sprintf("%s:%d", config.Proxy.Host, config.Proxy.Port),
						User:   url.UserPassword(config.Proxy.Username, config.Proxy.Password),
					}),
				},
			}
		case "socks5":
			dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("%s:%d", config.Proxy.Host, config.Proxy.Port), &proxy.Auth{
				User:     config.Proxy.Username,
				Password: config.Proxy.Password,
			}, proxy.Direct)
			if err != nil {
				return fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
			}
			HttpClient = &http.Client{
				Timeout: 5 * time.Second,
				Transport: &http.Transport{
					MaxIdleConnsPerHost: 10,
					Dial:                dialer.Dial,
				},
			}
		default:
			log.Fatalf("Unsupported proxy type: %s", config.Proxy.Type)
		}
	} else {
		HttpClient = &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 10,
			},
		}
	}

	var snapshotBand C.BandData
	var snapshotCards C.CardData
	var snapshotCharacters C.CharacterData
	var snapshotEvents C.EventsData
	var snapshotRecent C.Recent
	var snapshotSkills C.SkillData
	var snapshotGacha C.GachaData
	var snapshotRates C.RatesData
	var snapshotSongs C.SongsData
	var snapshotComic C.ComicData

	data := []fetchTask{
		{URL: "https://bestdori.com/api/bands/all.1.json", Target: &snapshotBand},
		{URL: "https://bestdori.com/api/cards/all.5.json", Target: &snapshotCards},
		{URL: "https://bestdori.com/api/characters/main.3.json", Target: &snapshotCharacters},
		{URL: "https://bestdori.com/api/events/all.5.json", Target: &snapshotEvents},
		{URL: "https://bestdori.com/api/news/dynamic/recent.json", Target: &snapshotRecent},
		{URL: "https://bestdori.com/api/skills/all.10.json", Target: &snapshotSkills},
		{URL: "https://bestdori.com/api/gacha/all.5.json", Target: &snapshotGacha},
		{URL: "https://bestdori.com/api/tracker/rates.json", Target: &snapshotRates},
		{URL: "https://bestdori.com/api/songs/all.5.json", Target: &snapshotSongs},
		{URL: "https://bestdori.com/api/comics/all.5.json", Target: &snapshotComic},
	}

	eg, gCtx := errgroup.WithContext(ctx)
	for _, task := range data {
		t := task
		eg.Go(func() error {
			return fetchAndDecode(gCtx, HttpClient, t.URL, t.Target)
		})
	}
	if err := eg.Wait(); err != nil {
		log.Printf("Error occurred while fetching and decoding data: %v", err)
		return err
	}

	publishDataSnapshot(snapshotBand, snapshotCards, snapshotCharacters, snapshotEvents, snapshotRecent, snapshotSkills, snapshotGacha, snapshotRates, snapshotSongs, snapshotComic)
	fmt.Println("All lists initialized successfully.")
	return nil
}

func fetchAndDecode(ctx context.Context, client *http.Client, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("Failed to create request [%s]: %w", url, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request [%s]: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected HTTP status code [%s]: %d", url, resp.StatusCode)
	}

	if err := json.UnmarshalRead(resp.Body, target); err != nil {
		return fmt.Errorf("Failed to decode JSON [%s]: %w", url, err)
	}

	return nil
}

func CronInit() {
	maxRetries := 5
	baseDelay := 30 * time.Second
	log.Printf("Starting CronInit with maxRetries=%d and baseDelay=%s", maxRetries, baseDelay)
	for i := range maxRetries {
		err := InitLists()
		if err == nil {
			log.Println("InitLists completed successfully.")
			break
		}
		delay := baseDelay * time.Duration(1<<i)
		if i == maxRetries-1 {
			log.Printf("InitLists failed after %d attempts: %v. No more retries.", maxRetries, err)
		} else {
			log.Printf("InitLists failed (attempt %d/%d): %v. Retrying in %s...", i+1, maxRetries, err, delay)
			time.Sleep(delay)
		}
	}
}
