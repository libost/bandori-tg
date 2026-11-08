package dynamic

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"sync"
	"time"

	C "github.com/libost/bandori-tg/constants"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/recent"
	"github.com/libost/bandori-tg/songs"
	"github.com/libost/bandori-tg/utils"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("dynamic", dynamicHandler))
	dispatcher.AddHandler(handlers.NewCommand("dyn", dynamicHandler))
	dispatcher.AddHandler(handlers.NewCommand("news", dynamicHandler))

}

func dynamicHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stopAction := make(chan struct{})
	var stopActionOnce sync.Once
	stopActionLoop := func() {
		stopActionOnce.Do(func() {
			close(stopAction)
		})
	}
	go func() {
		_, _ = b.SendChatAction(ctx.EffectiveChat.Id, "typing", nil)
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_, _ = b.SendChatAction(ctx.EffectiveChat.Id, "typing", nil)
			case <-stopAction:
				return
			}
		}
	}()
	defer stopActionLoop()
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, "en")
	regionCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, "jp")
	tz := utils.TzHelper(regionCode)
	loc, _ := time.LoadLocation(tz)
	if len(ctx.Args()) > 1 {
		if slices.Contains(C.AcceptedRegions, ctx.Args()[1]) {
			regionCode = ctx.Args()[1]
		}
	}
	idx := utils.IndexHelper(regionCode)
	recentMap := utils.ReadRecent()
	richMessage := &gotgbot.InputRichMessage{
		Blocks: []gotgbot.InputRichBlock{},
	}
	// Songs Section
	var recentSongs []string
	for songId, song := range recentMap.Songs {
		if shouldCollectSong(song.PublishedAt, idx) {
			recentSongs = append(recentSongs, songId)
		}
	}
	songMap := utils.ReadSongs()
	richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockSectionHeading{
		Text: gotgbot.RichTextString(I.GetLocalisedString("dynamic.songsHeading", langCode)),
		Size: 4,
	})
	if len(recentSongs) > 0 {
		var richMessageSongsSections []gotgbot.InputRichBlock
		for _, songId := range recentSongs {
			song := songMap[songId]
			pubTime, err := strconv.ParseInt(song.PublishedAt[idx], 10, 64)
			if err != nil {
				log.Printf("Error parsing published time for song %s: %v", songId, err)
				continue
			}
			pubTime = utils.TimeCheck(pubTime/1000) * 1000
			pubStr := time.Unix(pubTime/1000, 0).In(loc).Format("2006-01-02 15:04:05 MST")
			cells := [][]gotgbot.RichBlockTableCell{
				{
					{
						Text:  gotgbot.RichTextString(I.GetLocalisedString("dynamic.songsName", langCode)),
						Align: "left",
					},
					{
						Text:  gotgbot.RichTextString(song.MusicTitle[idx]),
						Align: "center",
					},
				},
				{
					{
						Text:  gotgbot.RichTextString(I.GetLocalisedString("dynamic.songsPubTime", langCode)),
						Align: "left",
					},
					{
						Text: gotgbot.RichTextDateTime{
							Text:           gotgbot.RichTextString(pubStr),
							UnixTime:       pubTime / 1000,
							DateTimeFormat: "DwT",
						},
						Align: "center",
					},
				},
			}
			thumbnailURL, _, err := songs.GetJacket(songId, song.JacketImage[0], regionCode)
			if err != nil {
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), "Error fetching song jacket: "+err.Error()), &gotgbot.SendMessageOpts{
					ParseMode: "HTML",
				})
				continue
			}
			richMessageSongsSections = append(richMessageSongsSections, gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByURL(thumbnailURL),
				},
			},
				gotgbot.InputRichBlockTable{
					Cells: cells,
				},
				gotgbot.InputRichBlockButtons{
					Buttons: []gotgbot.RichMessageButton{
						{
							Text:         gotgbot.RichTextString(I.GetLocalisedString("dynamic.songsButton", langCode)),
							CallbackData: "song_" + songId + "_" + regionCode,
							Style:        "primary",
						},
					},
				},
			)
		}
		richMessageSongDetails := gotgbot.InputRichBlockDetails{
			Blocks: richMessageSongsSections,
		}
		richMessage.Blocks = append(richMessage.Blocks, richMessageSongDetails)
	} else {
		richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockParagraph{
			Text: gotgbot.RichTextString(I.GetLocalisedString("dynamic.songsNone", langCode)),
		})
	}
	// Events Section
	richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockDivider{})
	richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockSectionHeading{
		Text: gotgbot.RichTextString(I.GetLocalisedString("dynamic.eventsHeading", langCode)),
		Size: 4,
	})
	recentEvents := recent.GetOngoingEventsID()
	if len(recentEvents) <= idx {
		richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockParagraph{
			Text: gotgbot.RichTextString(I.GetLocalisedString("dynamic.eventsNone", langCode)),
		})
	} else {
		if recentEvents[idx] != "" {
			eventsMap := utils.ReadEvents()
			event := eventsMap[recentEvents[idx]]
			endTime, err := strconv.ParseInt(event.EndAt[idx], 10, 64)
			if err != nil {
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), "Error parsing event end time: "+err.Error()), &gotgbot.SendMessageOpts{
					ParseMode: "HTML",
				})
				return err
			}
			endTime = utils.TimeCheck(endTime/1000) * 1000
			endStr := time.Unix(endTime/1000, 0).In(loc).Format("2006-01-02 15:04:05 MST")
			richMessageEventSections := []gotgbot.InputRichBlock{
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL("https://bestdori.com/assets/" + regionCode + "/event/" + event.AssetBundleName + "/images_rip/banner.png"),
					},
				},
				gotgbot.InputRichBlockTable{
					Cells: [][]gotgbot.RichBlockTableCell{
						{
							{
								Text:  gotgbot.RichTextString(I.GetLocalisedString("dynamic.eventsName", langCode)),
								Align: "left",
							},
							{
								Text:  gotgbot.RichTextString(event.EventName[idx]),
								Align: "center",
							},
						},
						{
							{
								Text:  gotgbot.RichTextString(I.GetLocalisedString("dynamic.eventsEndTime", langCode)),
								Align: "left",
							},
							{
								Text: gotgbot.RichTextDateTime{
									Text:           gotgbot.RichTextString(endStr),
									UnixTime:       endTime / 1000,
									DateTimeFormat: "DwT",
								},
								Align: "center",
							},
						},
					},
				},
				gotgbot.InputRichBlockButtons{
					Buttons: []gotgbot.RichMessageButton{
						{
							Text:         gotgbot.RichTextString(I.GetLocalisedString("dynamic.eventsButton", langCode)),
							CallbackData: "events_" + recentEvents[idx] + "_" + regionCode + "_" + strconv.FormatInt(time.Now().Unix(), 10),
							Style:        "primary",
						},
					},
				},
			}
			richMessageEventDetails := gotgbot.InputRichBlockDetails{
				Blocks: richMessageEventSections,
			}
			richMessage.Blocks = append(richMessage.Blocks, richMessageEventDetails)
		} else {
			richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockParagraph{
				Text: gotgbot.RichTextString(I.GetLocalisedString("dynamic.eventsNone", langCode)),
			})
		}
	}
	// Gacha Section
	richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockDivider{})
	richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockSectionHeading{
		Text: gotgbot.RichTextString(I.GetLocalisedString("dynamic.gachasHeading", langCode)),
		Size: 4,
	})

	recentGachas := recent.GetOngoingGachas()
	if len(recentGachas) <= idx {
		richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockParagraph{
			Text: gotgbot.RichTextString(I.GetLocalisedString("dynamic.gachasNone", langCode)),
		})
	} else {
		if len(recentGachas[idx].Gacha) > 0 {
			gachaMap := utils.ReadGacha()
			var richMessageGachaSections []gotgbot.InputRichBlock
			for _, gacha := range recentGachas[idx].Gacha {
				gachaDetails := gachaMap[gacha]
				richMessageGachaSections = append(richMessageGachaSections, gotgbot.InputRichBlockButtons{
					Buttons: []gotgbot.RichMessageButton{
						{
							Text:         gotgbot.RichTextString(gachaDetails.GachaName[idx]),
							CallbackData: fmt.Sprintf("gacha_%s_%d", gacha, idx),
							Style:        "primary",
						},
					},
				})
			}
			richMessageGachaDetails := gotgbot.InputRichBlockDetails{
				Blocks: richMessageGachaSections,
			}
			richMessage.Blocks = append(richMessage.Blocks, richMessageGachaDetails)
		} else {
			richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockParagraph{
				Text: gotgbot.RichTextString(I.GetLocalisedString("dynamic.gachasNone", langCode)),
			})
		}
	}
	// Login Bonus Section
	// Do we really need this?
	// We may reenable this if really needed, but for now, let's keep it commented out to avoid cluttering the message.
	/*
		richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockDivider{})
		richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockSectionHeading{
			Text: gotgbot.RichTextString(I.GetLocalisedString("dynamic.loginBonusHeading", langCode)),
			Size: 4,
		})
	*/
	_, err := b.SendRichMessageWithContext(ctxTimeout, ctx.EffectiveChat.Id, *richMessage, &gotgbot.SendRichMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	return err
}
