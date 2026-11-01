package characters

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/libost/bandori-tg/band"
	C "github.com/libost/bandori-tg/constants"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
)

func indexHelper(langCode string) int {
	switch langCode {
	case "jp":
		return 0
	case "en":
		return 1
	case "tw":
		return 2
	case "cn":
		return 3
	case "kr":
		return 4
	}
	return 0
}

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("character", characterCommand))
	dispatcher.AddHandler(handlers.NewCommand("char", characterCommand))
}

func characterCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	// Implementation of the character command goes here
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
	if len(ctx.Args()) == 1 {
		// TODO
		return fmt.Errorf("please provide a character ID")
	}
	characterId := ctx.Args()[1]
	idNum, err := strconv.Atoi(characterId)
	if err != nil {
		return fmt.Errorf("failed to convert character ID to number: %w", err)
	}
	if idNum < 1 || idNum > 40 {
		return fmt.Errorf("character ID must be between 1 and 100")
	}
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	qlangCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, "jp")
	qidx := indexHelper(qlangCode)
	url := "https://bestdori.com/assets/" + qlangCode + "/ui/character_kv_image/" + fmt.Sprintf("%03d", idNum) + "_rip/image.png"
	characterDetails, err := getCharacterDetails(characterId)
	if err != nil {
		return fmt.Errorf("failed to get character details: %w", err)
	}
	bandData, err := band.GetBand(fmt.Sprintf("%d", characterDetails.BandID))
	if err != nil {
		return fmt.Errorf("failed to get band data: %w", err)
	}
	schoolYear := convertAnyToString(characterDetails.Profile.SchoolYear[qidx])
	birthdayInt, err := strconv.Atoi(characterDetails.Profile.Birthday)
	if err != nil {
		return fmt.Errorf("failed to convert birthday to int: %w", err)
	}
	loc, _ := time.LoadLocation("Asia/Tokyo")
	bdStr := time.UnixMilli(int64(birthdayInt)).In(loc).Format("01-02")
	richMessage := gotgbot.InputRichMessage{
		Blocks: []gotgbot.InputRichBlock{
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(characterDetails.CharacterName[qidx]),
				Size: 5,
			},
			gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByURL(url),
				},
			},
			gotgbot.InputRichBlockTable{
				Cells: [][]gotgbot.RichBlockTableCell{
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.name", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(characterDetails.CharacterName[qidx]),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.voice", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(characterDetails.Profile.CharacterVoice[qidx]),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.color", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(characterDetails.ColorCode),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.band", langCode)),
							Align: "left",
						},
						{
							Text: gotgbot.RichTextArray{
								gotgbot.RichTextString(bandData.BandName[qidx] + " "),
								gotgbot.RichTextCustomEmoji{
									CustomEmojiId:   fmt.Sprintf("%d", C.BandEmoji[characterDetails.BandID-1]),
									AlternativeText: "",
								},
							},
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.role", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character."+characterDetails.Profile.Part, langCode)),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.birthday", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(bdStr),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.constellation", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character."+characterDetails.Profile.Constellation, langCode)),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.height", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(fmt.Sprintf("%dcm", characterDetails.Profile.Height)),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.school", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(characterDetails.Profile.School[qidx]),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.yearclass", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(schoolYear + " - " + characterDetails.Profile.SchoolCls[qidx]),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.favoritefood", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(characterDetails.Profile.FavoriteFood[qidx]),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.hatedfood", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(characterDetails.Profile.HatedFood[qidx]),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.hobby", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(characterDetails.Profile.Hobby[qidx]),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.selfintroduction", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(characterDetails.Profile.SelfIntroduction[qidx]),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("character.id", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(characterId),
							Align: "right",
						},
					},
				},
			},
		},
	}
	_, err = b.SendRichMessage(ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	return err
}
