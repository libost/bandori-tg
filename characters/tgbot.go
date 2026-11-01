package characters

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"strconv"
	"sync"
	"time"

	"github.com/libost/bandori-tg/band"
	C "github.com/libost/bandori-tg/constants"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/utils"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("character", characterCommand))
	dispatcher.AddHandler(handlers.NewCommand("char", characterCommand))
}

func characterCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	if len(ctx.Args()) == 1 {
		// TODO
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("character.usage", I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)), nil)
		return fmt.Errorf("no character parameter provided")
	}
	characterId := ctx.Args()[1]
	idNum, err := strconv.Atoi(characterId)
	if err != nil {
		index := C.NewSearchIndex(C.CharaSearch)
		matches := index.Search(characterId)
		if len(matches) == 0 {
			return fmt.Errorf("invalid character ID or name: %s", characterId)
		}
		idNum = matches[0].Item.ID
		characterId = fmt.Sprintf("%d", idNum)
	}
	if idNum < 1 || idNum > 40 {
		return fmt.Errorf("character ID invalid, expected to be between 1 and 40, got %d", idNum)
	}
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	qlangCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, "jp")
	qidx := utils.IndexHelper(qlangCode)
	characterDetails, err := getCharacterDetails(characterId)
	if err != nil {
		return fmt.Errorf("failed to get character details for %s: %w", characterId, err)
	}
	if characterDetails.CharacterName[qidx] == "" {
		for i, name := range characterDetails.CharacterName {
			if name != "" {
				qidx = i
				break
			}
		}
	}
	qlangCode = utils.ReverseIndexHelper(qidx)
	bandData, err := band.GetBand(fmt.Sprintf("%d", characterDetails.BandID))
	if err != nil {
		return fmt.Errorf("failed to get band data for %s: %w", characterId, err)
	}
	schoolYear := convertAnyToString(characterDetails.Profile.SchoolYear[qidx])
	birthdayInt, err := strconv.Atoi(characterDetails.Profile.Birthday)
	if err != nil {
		return fmt.Errorf("failed to convert birthday for %s to int: %w", characterId, err)
	}
	loc, _ := time.LoadLocation("Asia/Tokyo")
	bdStr := time.UnixMilli(int64(birthdayInt)).In(loc).Format("01-02")
	url := "https://bestdori.com/assets/jp/ui/character_kv_image/" + fmt.Sprintf("%03d", idNum) + "_rip/image.png"
	img, err := generateCharacterBackground(url, characterDetails.ColorCode, ctxTimeout)
	if err != nil {
		return fmt.Errorf("failed to generate character background for %s: %w", characterId, err)
	}
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return fmt.Errorf("failed to encode image for %s: %w", characterId, err)
	}
	richMessage := gotgbot.InputRichMessage{
		Blocks: []gotgbot.InputRichBlock{
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(characterDetails.CharacterName[qidx]),
				Size: 5,
			},
			gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByReader("image.png", &buf),
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
							Text:  gotgbot.RichTextString(fmt.Sprintf("%d cm", characterDetails.Profile.Height)),
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
	_, err = b.SendRichMessageWithContext(ctxTimeout, ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	return err
}
