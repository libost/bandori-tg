package characters

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"image/png"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/libost/bandori-tg/band"
	C "github.com/libost/bandori-tg/constants"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/utils"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
)

//go:embed signatures/*.png
var signatureFS embed.FS

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
		var queryString strings.Builder
		for i := range len(ctx.Args()) - 1 {
			queryString.WriteString(ctx.Args()[i+1])
			if i < len(ctx.Args())-2 {
				queryString.WriteString(" ")
			}
		}
		matches := index.Search(queryString.String())
		if len(matches) == 0 {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("character.character_not_found", I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)), nil)
			return fmt.Errorf("invalid character ID or name: %s", queryString.String())
		}
		idNum = matches[0].Item.ID
		characterId = fmt.Sprintf("%d", idNum)
	}
	if idNum < 1 || idNum > 40 {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("character.character_not_found", I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)), nil)
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
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("character.error_occurred", langCode), nil)
		return fmt.Errorf("failed to convert birthday for %s to int: %w", characterId, err)
	}
	loc, _ := time.LoadLocation("Asia/Tokyo")
	bdTimeStr := time.UnixMilli(int64(birthdayInt)).In(loc).Format("01-02")
	bdParts := strings.Split(bdTimeStr, "-")
	for i, part := range bdParts {
		bdParts[i] = strings.TrimPrefix(part, "0")
	}
	if langCode == "en" {
		// Convert month and day to English format
		month, err := strconv.Atoi(bdParts[0])
		if err != nil {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("character.error_occurred", langCode), nil)
			return fmt.Errorf("failed to convert birthday month for %s to int: %w", characterId, err)
		}
		bdParts[0] = time.Month(month).String()
		day, err := strconv.Atoi(bdParts[1])
		if err != nil {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("character.error_occurred", langCode), nil)
			return fmt.Errorf("failed to convert birthday day for %s to int: %w", characterId, err)
		}
		bdParts[1] = humanize.Ordinal(day)
	}
	bdStr := fmt.Sprintf(I.GetLocalisedString("character.birthday_format", langCode), bdParts[0], bdParts[1])
	url := "https://bestdori.com/assets/jp/ui/character_kv_image/" + fmt.Sprintf("%03d", idNum) + "_rip/image.png"
	img, err := generateCharacterBackground(url, characterDetails.ColorCode, ctxTimeout)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("character.error_occurred", langCode), nil)
		return fmt.Errorf("failed to generate character background for %s: %w", characterId, err)
	}
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("character.error_occurred", langCode), nil)
		return fmt.Errorf("failed to encode image for %s: %w", characterId, err)
	}
	signatureImg, err := signatureFS.ReadFile("signatures/" + characterDetails.LastName[1] + "_" + characterDetails.FirstName[1] + "_Signature.png")
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("character.error_occurred", langCode), nil)
		return fmt.Errorf("failed to read signature image for %s: %w", characterId, err)
	}
	richMessage := gotgbot.InputRichMessage{
		Blocks: []gotgbot.InputRichBlock{
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(characterDetails.CharacterName[qidx]),
				Size: 3,
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
							Text: gotgbot.RichTextArray{
								gotgbot.RichTextString(characterDetails.CharacterName[qidx] + " "),
								gotgbot.RichTextCustomEmoji{
									CustomEmojiId:   fmt.Sprintf("%d", C.CharaEmoji[idNum-1]),
									AlternativeText: "",
								},
							},
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
			gotgbot.InputRichBlockDivider{},
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(I.GetLocalisedString("character.signature", langCode)),
				Size: 6,
			},
			gotgbot.InputRichBlockDetails{
				Blocks: []gotgbot.InputRichBlock{
					gotgbot.InputRichBlockPhoto{
						Photo: gotgbot.InputMediaPhoto{
							Media: gotgbot.InputFileByReader("signature.png", bytes.NewReader(signatureImg)),
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
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("character.error_occurred", langCode), nil)
	}
	return err
}
