package utils

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libost/bandori-tg/config"
	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/version"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/kanjiconv"
	"golang.org/x/net/proxy"
	"golang.org/x/sync/errgroup"
)

var HttpClient *http.Client

type fetchTask struct {
	URL    string
	Target any
}

const (
	dataBand = iota
	dataCards
	dataCharacters
	dataEvents
	dataRecent
	dataSkills
	dataGacha
	dataRates
	dataSongs
	dataComic
	dataSongsMeta
)

var etags = make([]string, 11)

var dataMu sync.RWMutex

var Band C.BandsMap
var Cards C.CardsMap
var Characters C.CharactersMap
var Events C.EventsMap
var Recent C.Recent
var Skills C.SkillsMap
var Gacha C.GachasMap
var Rates C.RatesMap
var Songs C.SongsMap
var Comic C.ComicsMap
var SongsMeta C.SongMeta
var SongsKeys []*Item
var REventsKeys []string
var RGachaKeys []string

func ReadBand() C.BandsMap {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Band
}

func ReadCards() C.CardsMap {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Cards
}

func ReadCharacters() C.CharactersMap {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Characters
}

func ReadEvents() C.EventsMap {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Events
}

func ReadRecent() C.Recent {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Recent
}

func ReadSkills() C.SkillsMap {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Skills
}

func ReadGacha() C.GachasMap {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Gacha
}

func ReadRates() C.RatesMap {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Rates
}

func ReadSongs() C.SongsMap {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Songs
}

func ReadComic() C.ComicsMap {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return Comic
}

// ReadSongsMeta returns a copy of the SongsMeta data structure.
//
// example usage:
//
//	var songID = "1" // Yes! BanG_Dream! in this case
//	var difficulty = "3" // Expert in this case
//	var skilllength = "7" // 7 seconds skill length in this case
//	meta := utils.ReadSongsMeta()
//	factors, ok := meta[songID][difficulty][skilllength]
//	if !ok {
//		fmt.Println("Song meta not found for the given parameters")
//	}
//	fmt.Println("Factors:", factors)
//	fmt.Println("Score Factor without Fever:", factors.ScoreFactor(false, 1.0, 1.1))
//	fmt.Println("Score Factor with Fever:", factors.ScoreFactor(true, 1.0, 1.1))
//	fmt.Println("Skill Reliance without Fever:", factors.SkillReliance(false))
//	fmt.Println("Skill Reliance with Fever:", factors.SkillReliance(true))
//	length := 120.0 // Example song length in seconds, you should get the actual length from the song data
//	efficiency := factors.Efficiency(true, 1.0, 1.1, length)
//	fmt.Println("Efficiency with Fever:", efficiency)
//
// For a more "close to bestdori" result, you should turn every result into a percentage and round it to the nearest integer, as bestdori does.
func ReadSongsMeta() C.SongMeta {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return SongsMeta
}

func ReadSongsKeys() []*Item {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return append([]*Item(nil), SongsKeys...)
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

func publishDataSnapshot(changed []bool, snapshotBand C.BandsMap, snapshotCards C.CardsMap, snapshotCharacters C.CharactersMap, snapshotEvents C.EventsMap, snapshotRecent C.Recent, snapshotSkills C.SkillsMap, snapshotGacha C.GachasMap, snapshotRates C.RatesMap, snapshotSongs C.SongsMap, snapshotSongsMeta C.SongMeta, snapshotComic C.ComicsMap) {
	dataMu.Lock()
	defer dataMu.Unlock()

	if changed[dataBand] {
		Band = snapshotBand
	}
	if changed[dataCards] {
		Cards = snapshotCards
	}
	if changed[dataCharacters] {
		Characters = snapshotCharacters
	}
	if changed[dataEvents] {
		Events = snapshotEvents
	}
	if changed[dataRecent] {
		Recent = snapshotRecent
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
	}

	if changed[dataSkills] {
		Skills = snapshotSkills
	}
	if changed[dataGacha] {
		Gacha = snapshotGacha
	}
	if changed[dataRates] {
		Rates = snapshotRates
	}
	if changed[dataSongs] {
		Songs = snapshotSongs
		SongsKeys = make([]*Item, 0, len(Songs))
		t := kanjiconv.InitTokenizer()
		for i, k := range Songs {
			idx, _ := strconv.Atoi(i)
			romanizedTitle := kanjiconv.ConvertJapaneseToRoman(t, k.MusicTitle[0], &kanjiconv.ConvertOptions{
				CleanNN: true,
			})
			SongsKeys = append(SongsKeys, &Item{
				ID:      idx,
				Name:    k.MusicTitle[0],
				Aliases: append([]string{romanizedTitle}, k.MusicTitle[1:]...),
			})
		}
	}
	if changed[dataComic] {
		Comic = snapshotComic
	}
	if changed[dataSongsMeta] {
		SongsMeta = snapshotSongsMeta
	}
}

func InitLists() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var snapshotBand C.BandsMap
	var snapshotCards C.CardsMap
	var snapshotCharacters C.CharactersMap
	var snapshotEvents C.EventsMap
	var snapshotRecent C.Recent
	var snapshotSkills C.SkillsMap
	var snapshotGacha C.GachasMap
	var snapshotRates C.RatesMap
	var snapshotSongs C.SongsMap
	var snapshotComic C.ComicsMap
	var snapshotSongsMeta C.SongMeta

	var data = []fetchTask{
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
		{URL: "https://bestdori.com/api/songs/meta/all.5.json", Target: &snapshotSongsMeta},
	}

	eg, gCtx := errgroup.WithContext(ctx)
	changed := make([]bool, len(data))
	for i := range data {
		taskIndex := i
		task := &data[i]
		eg.Go(func() error {
			notModified, etag, err := fetchAndDecode(gCtx, HttpClient, *task, taskIndex)
			if err != nil {
				return err
			}
			if etag != "" {
				etags[taskIndex] = etag
			}
			if !notModified {
				changed[taskIndex] = true
			}
			return err
		})
	}
	if err := eg.Wait(); err != nil {
		log.Printf("Error occurred while fetching and decoding data: %v", err)
		return err
	}
	for _, sourceChanged := range changed {
		if sourceChanged {
			publishDataSnapshot(changed, snapshotBand, snapshotCards, snapshotCharacters, snapshotEvents, snapshotRecent, snapshotSkills, snapshotGacha, snapshotRates, snapshotSongs, snapshotSongsMeta, snapshotComic)
			break
		}
	}
	log.Printf("All lists initialized successfully.")
	return nil
}

func fetchAndDecode(ctx context.Context, client *http.Client, fetchTask fetchTask, etagIndex int) (bool, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchTask.URL, nil)
	if err != nil {
		return false, "", fmt.Errorf("failed to create request [%s]: %w", fetchTask.URL, err)
	}
	if etags[etagIndex] != "" {
		req.Header.Set("If-None-Match", etags[etagIndex])
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to send request [%s]: %w", fetchTask.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return true, "", nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("unexpected HTTP status code [%s]: %d", fetchTask.URL, resp.StatusCode)
	}

	if err := json.UnmarshalRead(resp.Body, fetchTask.Target); err != nil {
		return false, "", fmt.Errorf("failed to decode JSON [%s]: %w", fetchTask.URL, err)
	}

	return false, resp.Header.Get("ETag"), nil
}

func InitHttpClient() error {
	ua := generateUserAgent()
	config := config.AppConfig
	if config.Proxy.Enabled {
		switch config.Proxy.Type {
		case "http":
			HttpClient = &http.Client{
				Timeout: 5 * time.Second,
				Transport: &UATransport{
					Base: &http.Transport{
						MaxIdleConnsPerHost: 10,
						DialContext:         http.DefaultClient.Transport.(*http.Transport).DialContext,
						Proxy: http.ProxyURL(&url.URL{
							Scheme: "http",
							Host:   fmt.Sprintf("%s:%d", config.Proxy.Host, config.Proxy.Port),
							User:   url.UserPassword(config.Proxy.Username, config.Proxy.Password),
						}),
					},
					UserAgent: ua,
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
				Transport: &UATransport{
					Base: &http.Transport{
						MaxIdleConnsPerHost: 10,
						DialContext:         dialer.(proxy.ContextDialer).DialContext,
					},
					UserAgent: ua,
				},
			}
		default:
			return fmt.Errorf("unsupported proxy type: %s", config.Proxy.Type)
		}
	} else {
		HttpClient = &http.Client{
			Timeout: 5 * time.Second,
			Transport: &UATransport{
				Base: &http.Transport{
					MaxIdleConnsPerHost: 10,
				},
				UserAgent: ua,
			},
		}
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

func IndexHelper(str string) int {
	switch str {
	case "jp", "powerful":
		return 0
	case "en", "cool":
		return 1
	case "tw", "happy":
		return 2
	case "cn", "pure":
		return 3
	case "kr":
		return 4
	default:
		return 0
	}
}

func ReverseIndexHelper(index int) string {
	switch index {
	case 0:
		return "jp"
	case 1:
		return "en"
	case 2:
		return "tw"
	case 3:
		return "cn"
	case 4:
		return "kr"
	}
	return "jp"
}

func TzHelper(region string) string {
	switch region {
	case "jp":
		return "Asia/Tokyo"
	case "en":
		return "UTC"
	case "tw":
		return "Asia/Taipei"
	case "cn":
		return "Asia/Shanghai"
	case "kr":
		return "Asia/Seoul"
	default:
		return "Asia/Tokyo"
	}
}

func (t *UATransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqCopy := req.Clone(req.Context())

	if reqCopy.Header.Get("User-Agent") == "" {
		reqCopy.Header.Set("User-Agent", t.UserAgent)
	}

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(reqCopy)
}

type UATransport struct {
	Base      http.RoundTripper
	UserAgent string
}

func IsValidPNGUrl(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false
	}

	resp, err := HttpClient.Do(req)

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	contentType := resp.Header.Get("Content-Type")
	return contentType == "image/png"
}

// TimeCheck checks if the given timestamp is within the acceptable range defined by ApiTimeMaxInterval.
//
// If the timestamp is greater than the current time plus ApiTimeMaxInterval, it returns DefaultTimeStamp.
//
// Otherwise, it returns the original timestamp.
//
// timestamp: The timestamp to be checked, in seconds since the Unix epoch.
func TimeCheck(timestamp int64) int64 {
	now := time.Now().Unix()
	if timestamp > now+C.ApiTimeMaxIntervalSec {
		return C.DefaultTimeStampSec
	}
	return timestamp
}

func IsSpecifiedError(err error, targetDescription string) bool {
	if err == nil {
		return false
	}
	if tgErr, ok := errors.AsType[*gotgbot.TelegramError](err); ok {
		if tgErr.Code == 400 &&
			tgErr.Method == "sendRichMessage" &&
			strings.Contains(tgErr.Description, targetDescription) {
			return true
		}
	}
	return false
}

func generateUserAgent() string {
	version := strings.TrimPrefix(version.Version, "v")
	if version == "dev" {
		loc, _ := time.LoadLocation("UTC")
		version = "0.0.0-" + time.Now().In(loc).Format("20060102150405")
	}
	return fmt.Sprintf("bandori-tg/%s (https://github.com/libost/bandori-tg)", version)
}
