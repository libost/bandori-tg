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
	regionCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, "jp")
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	if len(ctx.Args()) == 1 {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.no_card_id", langCode), nil)
		stopActionLoop()
		log.Printf("User %d did not provide a card ID for query command.", ctx.EffectiveUser.Id)
		return nil
	}
	param := ctx.Args()[1]
	if len(ctx.Args()) > 2 {
		regionCode = ctx.Args()[2]
		if !slices.Contains(C.AcceptedRegions, regionCode) {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.invalid_region", langCode), nil)
			log.Printf("User %d provided an invalid region code: %s.", ctx.EffectiveUser.Id, regionCode)
			return nil
		}
	}
	// Process the card ID and query the database
	cardMap := utils.ReadCards()
	card, exists := cardMap[param]
	if !exists {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.card_not_found", langCode), nil)
		log.Printf("User %d queried a non-existent card with ID %s.", ctx.EffectiveUser.Id, param)
		return nil
	}
	// Process the card and send the response
	ridx := utils.IndexHelper(regionCode)
	regionCodeRetrieved := regionCodeFromCard(card, regionCode)
	if regionCode != regionCodeRetrieved {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("cards.region_no_data", langCode), regionCode, regionCodeRetrieved), nil)
		_, _ = b.SendChatAction(ctx.EffectiveChat.Id, "typing", nil)
	}
	regionCode = regionCodeRetrieved
	cardDetailed, err := GetDetailedCard(param)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", langCode), nil)
		log.Printf("Error fetching detailed card data for card ID %s: %v", param, err)
		return nil
	}
	gachaText := strings.ReplaceAll(selectLocaleString(cardDetailed.GachaText, ridx), "\n", "")
	cardType := I.GetLocalisedString("cards."+card.Type, langCode)
	name := selectLocaleString(card.Prefix, ridx)

	char, err := characters.GetCharacter(fmt.Sprint(card.CharacterID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", langCode), nil)
		log.Printf("Error fetching character data for card ID %s: %v", param, err)
		return nil
	}
	characterName := selectLocaleString(char.CharacterName, ridx)
	band, err := band.GetBand(fmt.Sprint(char.BandID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", langCode), nil)
		log.Printf("Error fetching band data for character ID %d: %v", card.CharacterID, err)
		return nil
	}
	bandName := selectLocaleString(band.BandName, ridx)
	normalPath, trainingPath, err := GetCard(param)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", langCode), nil)
		log.Printf("Error fetching card data for card ID %s: %v", param, err)
		return nil
	}
	rarityEmoji := fmt.Sprint(C.RarityEmoji[0])
	if trainingPath != "" {
		rarityEmoji = fmt.Sprint(C.RarityEmoji[1])
	}
	var rarityBlock gotgbot.RichTextArray
	for range card.Rarity {
		rarityBlock = append(rarityBlock, gotgbot.RichTextCustomEmoji{
			CustomEmojiId:   fmt.Sprint(rarityEmoji),
			AlternativeText: "😐",
		})
	}
	skill, err := skills.GetSkill(fmt.Sprint(card.SkillID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", langCode), nil)
		log.Printf("Error fetching skill data for card ID %s: %v", param, err)
		return nil
	}
	skillName := strings.ReplaceAll(selectLocaleString(skill.SimpleDescription, ridx), "\n", "")
	attribute := card.Attribute
	aidx := utils.IndexHelper(attribute)
	attribute = strings.ToUpper(attribute)
	releasedAt := selectLocaleString(card.ReleasedAt, ridx)
	releasedAtInt, _ := strconv.ParseInt(releasedAt, 10, 64)
	releasedAtInt = utils.TimeCheck(releasedAtInt/1000) * 1000 // Convert to milliseconds and check
	tz := utils.TzHelper(regionCode)
	loc, _ := time.LoadLocation(tz)
	releasedAtTime := time.Unix(releasedAtInt/1000, 0).In(loc).Format("2006-01-02 15:04:05 MST")
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
			Str:  gotgbot.RichTextString(I.GetLocalisedString(fmt.Sprintf("cards.%s", regionCode), langCode)),
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
					CustomEmojiId:   fmt.Sprint(C.AttributeEmoji[aidx]),
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
				Text:  gotgbot.RichTextString(I.GetLocalisedString("cards."+cell.Type, langCode)),
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
				Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_gacha_text", langCode)),
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
			Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_released_at", langCode)),
			Align: "left",
		},
		{
			Text: gotgbot.RichTextDateTime{
				Text:           gotgbot.RichTextString(releasedAtTime),
				UnixTime:       releasedAtInt / 1000,
				DateTimeFormat: "DwT",
			},
			Align: "right",
		},
	})
	richTable.Cells = append(richTable.Cells, []gotgbot.RichBlockTableCell{
		{
			Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_id", langCode)),
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
						Text: gotgbot.RichTextString(I.GetLocalisedString("cards.TrainingCard", langCode)),
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
						Text: gotgbot.RichTextString(I.GetLocalisedString("cards.NormalCard", langCode)),
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
						Text: gotgbot.RichTextString(I.GetLocalisedString("cards.NormalCard", langCode)),
					},
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(trainingPath),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(I.GetLocalisedString("cards.TrainingCard", langCode)),
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
		richMessageErr := &gotgbot.InputRichMessage{
			Blocks: []gotgbot.InputRichBlock{
				gotgbot.InputRichBlockSectionHeading{
					Text: gotgbot.RichTextString(name),
					Size: 3,
				},
			},
		}
		var richMessageErrPhotos []gotgbot.InputRichBlock
		if normalPath == "" {
			req, errReq := http.NewRequestWithContext(ctxTimeout, http.MethodGet, trainingPath, nil)
			if errReq != nil {
				log.Printf("Error creating request for training path: %v", errReq)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errReq.Error()), nil)
				return errReq
			}
			resp, errResp := http.DefaultClient.Do(req)
			if errResp != nil {
				log.Printf("Error fetching training path: %v", errResp)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errResp.Error()), nil)
				return errResp
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Printf("Training path returned non-OK status: %d", resp.StatusCode)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), fmt.Sprintf("Training path returned non-OK status: %d", resp.StatusCode)), nil)
				return fmt.Errorf("training path returned non-OK status: %d", resp.StatusCode)
			}
			var trainingPNG bytes.Buffer
			_, errCopy := io.Copy(&trainingPNG, resp.Body)
			if errCopy != nil {
				log.Printf("Error copying training PNG: %v", errCopy)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errCopy.Error()), nil)
				return errCopy
			}

			richMessageErrPhotos = append(richMessageErrPhotos, gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByReader("training.png", &trainingPNG),
				},
				Caption: &gotgbot.RichBlockCaption{
					Text: gotgbot.RichTextString(I.GetLocalisedString("cards.TrainingCard", langCode)),
				},
			})

		} else if trainingPath == "" {
			req, errReq := http.NewRequestWithContext(ctxTimeout, http.MethodGet, normalPath, nil)
			if errReq != nil {
				log.Printf("Error creating request for normal path: %v", errReq)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errReq.Error()), nil)
				return errReq
			}
			resp, errResp := http.DefaultClient.Do(req)
			if errResp != nil {
				log.Printf("Error fetching normal path: %v", errResp)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errResp.Error()), nil)
				return errResp
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Printf("Normal path returned non-OK status: %d", resp.StatusCode)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), fmt.Sprintf("Normal path returned non-OK status: %d", resp.StatusCode)), nil)
				return fmt.Errorf("normal path returned non-OK status: %d", resp.StatusCode)
			}
			var normalPNG bytes.Buffer
			_, errCopy := io.Copy(&normalPNG, resp.Body)
			if errCopy != nil {
				log.Printf("Error copying normal PNG: %v", errCopy)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errCopy.Error()), nil)
				return errCopy
			}
			richMessageErrPhotos = append(richMessageErrPhotos, gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByReader("normal.png", &normalPNG),
				},
				Caption: &gotgbot.RichBlockCaption{
					Text: gotgbot.RichTextString(I.GetLocalisedString("cards.NormalCard", langCode)),
				},
			})
		} else {
			reqTraining, errReqTraining := http.NewRequestWithContext(ctxTimeout, http.MethodGet, trainingPath, nil)
			if errReqTraining != nil {
				log.Printf("Error creating request for training path: %v", errReqTraining)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errReqTraining.Error()), nil)
				return errReqTraining
			}
			respTraining, errRespTraining := http.DefaultClient.Do(reqTraining)
			if errRespTraining != nil {
				log.Printf("Error fetching training path: %v", errRespTraining)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errRespTraining.Error()), nil)
				return errRespTraining
			}
			defer respTraining.Body.Close()
			if respTraining.StatusCode != http.StatusOK {
				log.Printf("Training path returned non-OK status: %d", respTraining.StatusCode)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), fmt.Sprintf("Training path returned non-OK status: %d", respTraining.StatusCode)), nil)
				return fmt.Errorf("training path returned non-OK status: %d", respTraining.StatusCode)
			}
			var trainingPNG bytes.Buffer
			_, errCopyTraining := io.Copy(&trainingPNG, respTraining.Body)
			if errCopyTraining != nil {
				log.Printf("Error copying training PNG: %v", errCopyTraining)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errCopyTraining.Error()), nil)
				return errCopyTraining
			}
			reqNormal, errReqNormal := http.NewRequestWithContext(ctxTimeout, http.MethodGet, normalPath, nil)
			if errReqNormal != nil {
				log.Printf("Error creating request for normal path: %v", errReqNormal)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errReqNormal.Error()), nil)
				return errReqNormal
			}
			respNormal, errRespNormal := http.DefaultClient.Do(reqNormal)
			if errRespNormal != nil {
				log.Printf("Error fetching normal path: %v", errRespNormal)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errRespNormal.Error()), nil)
				return errRespNormal
			}
			defer respNormal.Body.Close()
			if respNormal.StatusCode != http.StatusOK {
				log.Printf("Normal path returned non-OK status: %d", respNormal.StatusCode)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), fmt.Sprintf("Normal path returned non-OK status: %d", respNormal.StatusCode)), nil)
				return fmt.Errorf("normal path returned non-OK status: %d", respNormal.StatusCode)
			}
			var normalPNG bytes.Buffer
			_, errCopyNormal := io.Copy(&normalPNG, respNormal.Body)
			if errCopyNormal != nil {
				log.Printf("Error copying normal PNG: %v", errCopyNormal)
				ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errCopyNormal.Error()), nil)
				return errCopyNormal
			}
			richMessageErrPhotos = append(richMessageErrPhotos, gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByReader("normal.png", &normalPNG),
				},
				Caption: &gotgbot.RichBlockCaption{
					Text: gotgbot.RichTextString(I.GetLocalisedString("cards.NormalCard", langCode)),
				},
			})
			richMessageErrPhotos = append(richMessageErrPhotos, gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByReader("training.png", &trainingPNG),
				},
				Caption: &gotgbot.RichBlockCaption{
					Text: gotgbot.RichTextString(I.GetLocalisedString("cards.TrainingCard", langCode)),
				},
			})
		}
		richMessageErr.Blocks = append(richMessageErr.Blocks, richMessageErrPhotos...)
		richMessageErr.Blocks = append(richMessageErr.Blocks, gotgbot.InputRichBlockTable{
			Cells: richTable.Cells,
		})
		_, errErr := b.SendRichMessageWithContext(ctxTimeout, ctx.EffectiveChat.Id, *richMessageErr, &gotgbot.SendRichMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: ctx.EffectiveMessage.MessageId,
			},
		})
		if errErr != nil {
			log.Printf("Failed to send rich message with card details after handling error: %v", errErr)
			ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), errErr.Error()), nil)
		}
		return errErr
	}

	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.request_timeout", langCode), nil)
		log.Printf("Request timed out while sending card details to user %d", ctx.EffectiveUser.Id)
	} else if err != nil {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), err.Error()), &gotgbot.SendMessageOpts{
			ParseMode: "HTML",
		})
	}
	return err
}
