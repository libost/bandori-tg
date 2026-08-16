package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	C "github.com/libost/bandori-tg/constants"
	"golang.org/x/sync/errgroup"
)

type fetchTask struct {
	URL    string
	Target any
}

var Band C.BandData
var Cards C.CardData
var Characters C.CharacterData
var Events C.EventsData
var Recent C.Recent
var Skills C.SkillData
var Gacha C.GachaData
var REventsKeys []string

var tasks = []fetchTask{
	{
		URL:    "https://bestdori.com/api/bands/main.1.json",
		Target: &Band,
	},
	{
		URL:    "https://bestdori.com/api/cards/all.5.json",
		Target: &Cards,
	},
	{
		URL:    "https://bestdori.com/api/characters/main.3.json",
		Target: &Characters,
	},
	{
		URL:    "https://bestdori.com/api/events/all.5.json",
		Target: &Events,
	},
	{
		URL:    "https://bestdori.com/api/news/dynamic/recent.json",
		Target: &Recent,
	},
	{
		URL:    "https://bestdori.com/api/skills/all.10.json",
		Target: &Skills,
	},
	{
		URL:    "https://bestdori.com/api/gacha/all.5.json",
		Target: &Gacha,
	},
}

func InitLists() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10, // 提升并发复用效率
		},
	}
	eg, gCtx := errgroup.WithContext(ctx)
	for _, task := range tasks {
		// 注意 Go 1.22 以前需要闭包变量捕获：t := task
		t := task

		eg.Go(func() error {
			return fetchAndDecode(gCtx, httpClient, t.URL, t.Target)
		})
	}
	if err := eg.Wait(); err != nil {
		log.Printf("Error occurred while fetching and decoding data: %v", err)
		return err
	}
	fmt.Println("All lists initialized successfully.")
	return nil
}

func fetchAndDecode(ctx context.Context, client *http.Client, url string, target any) error {
	// 创建带 Context 的 Request，支持中途取消或超时
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("Failed to create request [%s]: %w", url, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request [%s]: %w", url, err)
	}
	defer resp.Body.Close()

	// 校验 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected HTTP status code [%s]: %d", url, resp.StatusCode)
	}

	// 使用 json.NewDecoder 直接从 Response Body 流式解码，高效省内存
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("Failed to decode JSON [%s]: %w", url, err)
	}
	if target == &Recent {
		REventsKeys = make([]string, 0, len(Recent.Events))
		for k := range Recent.Events {
			REventsKeys = append(REventsKeys, k)
		}
		sort.Strings(REventsKeys)
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
		delay := baseDelay * time.Duration(1<<i) // 指数退避
		if i == maxRetries-1 {
			log.Printf("InitLists failed after %d attempts: %v. No more retries.", maxRetries, err)
		} else {
			log.Printf("InitLists failed (attempt %d/%d): %v. Retrying in %s...", i+1, maxRetries, err, delay)
		}
		time.Sleep(delay)
	}
}
