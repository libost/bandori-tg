package utils

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/libost/bandori-tg/config"
	C "github.com/libost/bandori-tg/constants"
	"golang.org/x/sync/errgroup"
)

//go:embed fonts/*.ttf
var FontFS embed.FS

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
var REventsKeys []string

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

func ReadREventsKeys() []string {
	dataMu.RLock()
	defer dataMu.RUnlock()
	return append([]string(nil), REventsKeys...)
}

func publishDataSnapshot(snapshotBand C.BandData, snapshotCards C.CardData, snapshotCharacters C.CharacterData, snapshotEvents C.EventsData, snapshotRecent C.Recent, snapshotSkills C.SkillData, snapshotGacha C.GachaData) {
	dataMu.Lock()
	defer dataMu.Unlock()

	Band = snapshotBand
	Cards = snapshotCards
	Characters = snapshotCharacters
	Events = snapshotEvents
	Recent = snapshotRecent
	Skills = snapshotSkills
	Gacha = snapshotGacha

	REventsKeys = make([]string, 0, len(Recent.Events))
	for k := range Recent.Events {
		REventsKeys = append(REventsKeys, k)
	}
	sort.Strings(REventsKeys)
}

func InitLists() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
		},
	}

	var snapshotBand C.BandData
	var snapshotCards C.CardData
	var snapshotCharacters C.CharacterData
	var snapshotEvents C.EventsData
	var snapshotRecent C.Recent
	var snapshotSkills C.SkillData
	var snapshotGacha C.GachaData

	data := []fetchTask{
		{URL: "https://bestdori.com/api/bands/main.1.json", Target: &snapshotBand},
		{URL: "https://bestdori.com/api/cards/all.5.json", Target: &snapshotCards},
		{URL: "https://bestdori.com/api/characters/main.3.json", Target: &snapshotCharacters},
		{URL: "https://bestdori.com/api/events/all.5.json", Target: &snapshotEvents},
		{URL: "https://bestdori.com/api/news/dynamic/recent.json", Target: &snapshotRecent},
		{URL: "https://bestdori.com/api/skills/all.10.json", Target: &snapshotSkills},
		{URL: "https://bestdori.com/api/gacha/all.5.json", Target: &snapshotGacha},
	}

	eg, gCtx := errgroup.WithContext(ctx)
	for _, task := range data {
		t := task
		eg.Go(func() error {
			return fetchAndDecode(gCtx, httpClient, t.URL, t.Target)
		})
	}
	if err := eg.Wait(); err != nil {
		log.Printf("Error occurred while fetching and decoding data: %v", err)
		return err
	}

	publishDataSnapshot(snapshotBand, snapshotCards, snapshotCharacters, snapshotEvents, snapshotRecent, snapshotSkills, snapshotGacha)
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

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
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

// 由于未知的原因，SendRichMessage中无法直接使用InputFileByReader上传图片，因此需要先上传图片到PicDepot，然后再使用file_id发送图片
func GetImageIDWorkAround(b *gotgbot.Bot, buf bytes.Buffer) (string, error) {
	picDepot := "@" + config.AppConfig.General.PicDepot
	inputFile := gotgbot.InputFileByReader("image.png", &buf)
	msg, err := b.SendPhoto(0, inputFile, &gotgbot.SendPhotoOpts{
		RequestOpts: &gotgbot.RequestOpts{
			OverrideParams: map[string]any{
				"chat_id": picDepot,
			},
		},
	})
	if err != nil {
		return "", err
	}
	if len(msg.Photo) == 0 {
		return "", fmt.Errorf("no photo returned from Telegram")
	}
	fileID := msg.Photo[len(msg.Photo)-1].FileId
	return fileID, nil
}
