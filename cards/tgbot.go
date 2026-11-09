package cards

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libost/bandori-tg/band"
	"github.com/libost/bandori-tg/characters"
	C "github.com/libost/bandori-tg/constants"
	DB "github.com/libost/bandori-tg/database"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/skills"
	"github.com/libost/bandori-tg/utils"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
)

// XMLNode 定义递归 XML 节点结构
type XMLNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Content  []byte     `xml:",chardata"`
	Children []XMLNode  `xml:",any"`
}

func selectLocaleString(strings []string, index int) string {
	var selected string
	if strings[index] != "" && strings[index] != "null" {
		selected = strings[index]
	} else {
		for _, str := range strings {
			if str != "null" {
				selected = str
				break
			}
		}
	}
	return selected
}

func isWebPageContentErr(err error) bool {
	if err == nil {
		return false
	}
	if tgErr, ok := errors.AsType[*gotgbot.TelegramError](err); ok {
		if tgErr.Code == 400 && strings.Contains(tgErr.Description, "wrong type of the web page content") {
			return true
		}
	}
	return false
}

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("cards", QueryHandler))
	dispatcher.AddHandler(handlers.NewCommand("card", QueryHandler))
}

func QueryHandler(b *gotgbot.Bot, ctx *ext.Context) error {
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
	DB.Init("create", ctx.EffectiveUser.Id, nil)
	qlangCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, "jp")
	dlangCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	if ctx.EffectiveChat.Type != "private" {
	}
	if len(ctx.Args()) == 1 {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.no_card_id", dlangCode), nil)
		stopActionLoop()
		log.Printf("User %d did not provide a card ID for query command.", ctx.EffectiveUser.Id)
		return nil
	}
	param := ctx.Args()[1]
	if len(ctx.Args()) > 2 {
		qlangCode = ctx.Args()[2]
		if !slices.Contains(C.AcceptedRegions, qlangCode) {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.invalid_region", dlangCode), nil)
			log.Printf("User %d provided an invalid region code: %s.", ctx.EffectiveUser.Id, qlangCode)
			return nil
		}
	}
	// Process the card ID and query the database
	cardMap := utils.ReadCards()
	card, exists := cardMap[param]
	if !exists {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.card_not_found", dlangCode), nil)
		log.Printf("User %d queried a non-existent card with ID %s.", ctx.EffectiveUser.Id, param)
		return nil
	}
	// Process the card and send the response
	index := utils.IndexHelper(qlangCode)
	regionCode := regionCodeFromCard(card, qlangCode)
	if regionCode != qlangCode {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("cards.region_no_data", dlangCode), qlangCode, regionCode), nil)
		_, _ = b.SendChatAction(ctx.EffectiveChat.Id, "typing", nil)
	}
	cardDetailed, err := GetDetailedCard(param)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", dlangCode), nil)
		log.Printf("Error fetching detailed card data for card ID %s: %v", param, err)
		return nil
	}
	gachaText := strings.ReplaceAll(selectLocaleString(cardDetailed.GachaText, index), "\n", "")
	cardType := I.GetLocalisedString("cards."+card.Type, dlangCode)
	name := selectLocaleString(card.Prefix, index)

	char, err := characters.GetCharacter(fmt.Sprint(card.CharacterID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", dlangCode), nil)
		log.Printf("Error fetching character data for card ID %s: %v", param, err)
		return nil
	}
	characterName := selectLocaleString(char.CharacterName, index)
	band, err := band.GetBand(fmt.Sprint(char.BandID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", dlangCode), nil)
		log.Printf("Error fetching band data for character ID %d: %v", card.CharacterID, err)
		return nil
	}
	bandName := selectLocaleString(band.BandName, index)
	normalPath, trainingPath, err := GetCard(param)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", dlangCode), nil)
		log.Printf("Error fetching card data for card ID %s: %v", param, err)
		return nil
	}
	rarityEmoji := fmt.Sprint(C.RarityEmoji[0])
	if trainingPath != "" {
		rarityEmoji = fmt.Sprint(C.RarityEmoji[1])
	}
	var rarityBlock gotgbot.RichTextArray
	for i := 0; i < card.Rarity; i++ {
		rarityBlock = append(rarityBlock, gotgbot.RichTextCustomEmoji{
			CustomEmojiId:   fmt.Sprint(rarityEmoji),
			AlternativeText: "😐",
		})
	}
	skill, err := skills.GetSkill(fmt.Sprint(card.SkillID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", dlangCode), nil)
		log.Printf("Error fetching skill data for card ID %s: %v", param, err)
		return nil
	}
	skillName := strings.ReplaceAll(selectLocaleString(skill.SimpleDescription, index), "\n", "")
	attribute := card.Attribute
	aindex := utils.IndexHelper(attribute)
	attribute = strings.ToUpper(attribute)
	releasedAt := selectLocaleString(card.ReleasedAt, index)
	releasedAtInt, _ := strconv.Atoi(releasedAt)
	releasedAtInt64 := utils.TimeCheck(int64(releasedAtInt)/1000) * 1000 // Convert to milliseconds and check
	tz := utils.TzHelper(qlangCode)
	loc, _ := time.LoadLocation(tz)
	releasedAtTime := time.Unix(releasedAtInt64/1000, 0).In(loc).Format("2006-01-02 15:04:05 MST")
	log.Printf("User %d queried card ID %s", ctx.EffectiveUser.Id, param)
	cemojiID := C.CharaEmoji[card.CharacterID-1]
	bemojiID := C.BandEmoji[char.BandID-1]

	type tableCell struct {
		Type string
		Str  gotgbot.RichText
	}
	cells := []tableCell{
		{
			Type: "region",
			Str:  gotgbot.RichTextString(I.GetLocalisedString(fmt.Sprintf("cards.%s", regionCode), dlangCode)),
		},
		{
			Type: "card_info_title",
			Str:  gotgbot.RichTextString(name),
		},
		{
			Type: "card_info_type",
			Str:  gotgbot.RichTextString(cardType),
		},
		{
			Type: "card_info_chara",
			Str: gotgbot.RichTextArray{
				gotgbot.RichTextString(characterName + " "),
				gotgbot.RichTextCustomEmoji{
					CustomEmojiId:   fmt.Sprint(cemojiID),
					AlternativeText: "😐",
				},
			},
		},
		{
			Type: "card_info_band",
			Str: gotgbot.RichTextArray{
				gotgbot.RichTextString(bandName + " "),
				gotgbot.RichTextCustomEmoji{
					CustomEmojiId:   fmt.Sprint(bemojiID),
					AlternativeText: "😐",
				},
			},
		},
		{
			Type: "card_info_attribute",
			Str: gotgbot.RichTextArray{
				gotgbot.RichTextString(attribute + " "),
				gotgbot.RichTextCustomEmoji{
					CustomEmojiId:   fmt.Sprint(C.AttributeEmoji[aindex]),
					AlternativeText: "😐",
				},
			},
		},
		{
			Type: "card_info_rarity",
			Str:  rarityBlock,
		},
		{
			Type: "card_info_skill",
			Str:  gotgbot.RichTextString(skillName),
		},
	}
	richCell := make([][]gotgbot.RichBlockTableCell, len(cells))
	for i, cell := range cells {
		richCell[i] = []gotgbot.RichBlockTableCell{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("cards."+cell.Type, dlangCode)),
				Align: "left",
			},
			{
				Text:  cell.Str,
				Align: "right",
			},
		}
	}

	richTable := &gotgbot.RichBlockTable{
		Cells: richCell,
	}
	if gachaText != "" {
		richTable.Cells = append(richTable.Cells, []gotgbot.RichBlockTableCell{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_gacha_text", dlangCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(gachaText),
				Align: "right",
			},
		})
	}
	richTable.Cells = append(richTable.Cells, []gotgbot.RichBlockTableCell{
		{
			Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_released_at", dlangCode)),
			Align: "left",
		},
		{
			Text: gotgbot.RichTextDateTime{
				Text:           gotgbot.RichTextString(releasedAtTime),
				UnixTime:       releasedAtInt64 / 1000,
				DateTimeFormat: "DwT",
			},
			Align: "right",
		},
	})
	richTable.Cells = append(richTable.Cells, []gotgbot.RichBlockTableCell{
		{
			Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_id", dlangCode)),
			Align: "left",
		},
		{
			Text:  gotgbot.RichTextString(param),
			Align: "right",
		},
	})
	var richMessage *gotgbot.InputRichMessage
	if normalPath == "" {
		richMessage = &gotgbot.InputRichMessage{
			Blocks: []gotgbot.InputRichBlock{
				gotgbot.InputRichBlockSectionHeading{
					Text: gotgbot.RichTextString(name),
					Size: 3,
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(trainingPath),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(I.GetLocalisedString("cards.TrainingCard", dlangCode)),
					},
				},
			},
		}
	} else if trainingPath == "" {
		richMessage = &gotgbot.InputRichMessage{
			Blocks: []gotgbot.InputRichBlock{
				gotgbot.InputRichBlockSectionHeading{
					Text: gotgbot.RichTextString(name),
					Size: 3,
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(normalPath),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(I.GetLocalisedString("cards.NormalCard", dlangCode)),
					},
				},
			},
		}
	} else {
		richMessage = &gotgbot.InputRichMessage{
			Blocks: []gotgbot.InputRichBlock{
				gotgbot.InputRichBlockSectionHeading{
					Text: gotgbot.RichTextString(name),
					Size: 3,
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(normalPath),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(I.GetLocalisedString("cards.NormalCard", dlangCode)),
					},
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(trainingPath),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(I.GetLocalisedString("cards.TrainingCard", dlangCode)),
					},
				},
			},
		}
	}
	richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockTable{
		Cells: richTable.Cells,
	})
	_, err = b.SendRichMessageWithContext(ctxTimeout, ctx.EffectiveChat.Id, *richMessage, &gotgbot.SendRichMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	if utils.IsSpecifiedError(err, "failed to get HTTP URL content") || utils.IsSpecifiedError(err, "wrong type of the web page content") {
		// Handle the specific error
		var richMessageErr *gotgbot.InputRichMessage
		if normalPath == "" {
			req, errReq := http.NewRequestWithContext(ctxTimeout, http.MethodGet, trainingPath, nil)
			if errReq != nil {
				log.Printf("Error creating request for training path: %v", errReq)
				return errReq
			}
			resp, errResp := http.DefaultClient.Do(req)
			if errResp != nil {
				log.Printf("Error fetching training path: %v", errResp)
				return errResp
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Printf("Training path returned non-OK status: %d", resp.StatusCode)
				return fmt.Errorf("training path returned non-OK status: %d", resp.StatusCode)
			}
			var trainingPNG bytes.Buffer
			_, errCopy := io.Copy(&trainingPNG, resp.Body)
			if errCopy != nil {
				log.Printf("Error copying training PNG: %v", errCopy)
				return errCopy
			}
			richMessageErr = &gotgbot.InputRichMessage{
				Blocks: []gotgbot.InputRichBlock{
					gotgbot.InputRichBlockSectionHeading{
						Text: gotgbot.RichTextString(name),
						Size: 3,
					},
					gotgbot.InputRichBlockPhoto{
						Photo: gotgbot.InputMediaPhoto{
							Media: gotgbot.InputFileByReader("training.png", &trainingPNG),
						},
						Caption: &gotgbot.RichBlockCaption{
							Text: gotgbot.RichTextString(I.GetLocalisedString("cards.TrainingCard", dlangCode)),
						},
					},
				},
			}
			_, errErr := b.SendRichMessageWithContext(ctxTimeout, ctx.EffectiveChat.Id, *richMessageErr, &gotgbot.SendRichMessageOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId: ctx.EffectiveMessage.MessageId,
				},
			})
			return errErr
		} else if trainingPath == "" {
			req, errReq := http.NewRequestWithContext(ctxTimeout, http.MethodGet, normalPath, nil)
			if errReq != nil {
				log.Printf("Error creating request for normal path: %v", errReq)
				return errReq
			}
			resp, errResp := http.DefaultClient.Do(req)
			if errResp != nil {
				log.Printf("Error fetching normal path: %v", errResp)
				return errResp
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Printf("Normal path returned non-OK status: %d", resp.StatusCode)
				return fmt.Errorf("normal path returned non-OK status: %d", resp.StatusCode)
			}
			var normalPNG bytes.Buffer
			_, errCopy := io.Copy(&normalPNG, resp.Body)
			if errCopy != nil {
				log.Printf("Error copying normal PNG: %v", errCopy)
				return errCopy
			}
			richMessageErr = &gotgbot.InputRichMessage{
				Blocks: []gotgbot.InputRichBlock{
					gotgbot.InputRichBlockSectionHeading{
						Text: gotgbot.RichTextString(name),
						Size: 3,
					},
					gotgbot.InputRichBlockPhoto{
						Photo: gotgbot.InputMediaPhoto{
							Media: gotgbot.InputFileByReader("normal.png", &normalPNG),
						},
						Caption: &gotgbot.RichBlockCaption{
							Text: gotgbot.RichTextString(I.GetLocalisedString("cards.NormalCard", dlangCode)),
						},
					},
				},
			}
			_, errErr := b.SendRichMessageWithContext(ctxTimeout, ctx.EffectiveChat.Id, *richMessageErr, &gotgbot.SendRichMessageOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId: ctx.EffectiveMessage.MessageId,
				},
			})
			return errErr
		} else {
			reqTraining, errReqTraining := http.NewRequestWithContext(ctxTimeout, http.MethodGet, trainingPath, nil)
			if errReqTraining != nil {
				log.Printf("Error creating request for training path: %v", errReqTraining)
				return errReqTraining
			}
			respTraining, errRespTraining := http.DefaultClient.Do(reqTraining)
			if errRespTraining != nil {
				log.Printf("Error fetching training path: %v", errRespTraining)
				return errRespTraining
			}
			defer respTraining.Body.Close()
			if respTraining.StatusCode != http.StatusOK {
				log.Printf("Training path returned non-OK status: %d", respTraining.StatusCode)
				return fmt.Errorf("training path returned non-OK status: %d", respTraining.StatusCode)
			}
			var trainingPNG bytes.Buffer
			_, errCopyTraining := io.Copy(&trainingPNG, respTraining.Body)
			if errCopyTraining != nil {
				log.Printf("Error copying training PNG: %v", errCopyTraining)
				return errCopyTraining
			}
			reqNormal, errReqNormal := http.NewRequestWithContext(ctxTimeout, http.MethodGet, normalPath, nil)
			if errReqNormal != nil {
				log.Printf("Error creating request for normal path: %v", errReqNormal)
				return errReqNormal
			}
			respNormal, errRespNormal := http.DefaultClient.Do(reqNormal)
			if errRespNormal != nil {
				log.Printf("Error fetching normal path: %v", errRespNormal)
				return errRespNormal
			}
			defer respNormal.Body.Close()
			if respNormal.StatusCode != http.StatusOK {
				log.Printf("Normal path returned non-OK status: %d", respNormal.StatusCode)
				return fmt.Errorf("normal path returned non-OK status: %d", respNormal.StatusCode)
			}
			var normalPNG bytes.Buffer
			_, errCopyNormal := io.Copy(&normalPNG, respNormal.Body)
			if errCopyNormal != nil {
				log.Printf("Error copying normal PNG: %v", errCopyNormal)
				return errCopyNormal
			}
			richMessageErr = &gotgbot.InputRichMessage{
				Blocks: []gotgbot.InputRichBlock{
					gotgbot.InputRichBlockSectionHeading{
						Text: gotgbot.RichTextString(name),
						Size: 3,
					},
					gotgbot.InputRichBlockPhoto{
						Photo: gotgbot.InputMediaPhoto{
							Media: gotgbot.InputFileByReader("normal.png", &normalPNG),
						},
						Caption: &gotgbot.RichBlockCaption{
							Text: gotgbot.RichTextString(I.GetLocalisedString("cards.NormalCard", dlangCode)),
						},
					},
					gotgbot.InputRichBlockPhoto{
						Photo: gotgbot.InputMediaPhoto{
							Media: gotgbot.InputFileByReader("training.png", &trainingPNG),
						},
						Caption: &gotgbot.RichBlockCaption{
							Text: gotgbot.RichTextString(I.GetLocalisedString("cards.TrainingCard", dlangCode)),
						},
					},
				},
			}
		}
		richMessageErr.Blocks = append(richMessageErr.Blocks, gotgbot.InputRichBlockTable{
			Cells: richTable.Cells,
		})
		_, errErr := b.SendRichMessageWithContext(ctxTimeout, ctx.EffectiveChat.Id, *richMessageErr, &gotgbot.SendRichMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: ctx.EffectiveMessage.MessageId,
			},
		})
		return errErr
	}

	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.request_timeout", dlangCode), nil)
		log.Printf("Request timed out while sending card details to user %d", ctx.EffectiveUser.Id)
	} else if err != nil {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", dlangCode), err.Error()), &gotgbot.SendMessageOpts{
			ParseMode: "HTML",
		})
	}
	return err
}
