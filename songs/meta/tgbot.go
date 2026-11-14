package meta

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	C "github.com/libost/bandori-tg/constants"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/songs"
	"github.com/libost/bandori-tg/utils"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("meta", metaCommand))
}

func metaCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	if len(ctx.Args()) < 2 {
		ctx.EffectiveMessage.Reply(b, "Please provide a song ID. Usage: /meta <song_id>", nil)
		return nil
	}
	regionCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, "jp")
	ridx := utils.IndexHelper(regionCode)
	songId := ctx.Args()[1]
	_, err := strconv.Atoi(songId)
	if err != nil {
		var searchStr strings.Builder
		for i, arg := range ctx.Args()[1:] {
			searchStr.WriteString(arg)
			if i < len(ctx.Args())-2 {
				searchStr.WriteString(" ")
			}
		}
		index := utils.NewSearchIndex(utils.SongsKeys)
		songs := index.Search(searchStr.String())
		if len(songs) == 0 {
			ctx.EffectiveMessage.Reply(b, fmt.Sprintf("No results found for '%s'.", searchStr.String()), nil)
			return nil
		}
		for _, song := range songs {
			if ridx == 0 {
				if song.Item.Name == "" {
					continue
				}
				songId = strconv.Itoa(song.Item.ID)
				break
			}
			if song.Item.Aliases[ridx-1] == "" {
				continue
			}
			songId = strconv.Itoa(song.Item.ID)
			break
		}
	}
	meta := utils.ReadSongsMeta()
	skilllength := "7" // Temporary hardcoded skill length for demonstration purposes

	songMap, err := songs.GetDetailedSong(songId)
	if err != nil {
		fmt.Printf("Error occurred while fetching songs: %v\n", err)
		return nil
	}
	difficultyMap := make([]difficultyData, len(songMap.Difficulty))
	for i, diff := range songMap.Difficulty {
		idx, _ := strconv.Atoi(i)
		difficultyMap[idx].PlayLevel = diff.PlayLevel
	}
	for i, bpm := range songMap.Bpm {
		idx, err := strconv.Atoi(i)
		if err != nil || idx < 0 || idx >= len(difficultyMap) {
			fmt.Printf("Invalid BPM difficulty index %q\n", i)
			continue
		}
		if len(bpm) == 0 {
			fmt.Printf("No BPM data for difficulty index %q\n", i)
			continue
		}
		difficultyMap[idx].Bpm = bpm[0].Bpm
	}
	for i, note := range songMap.Notes {
		idx, err := strconv.Atoi(i)
		if err != nil || idx < 0 || idx >= len(difficultyMap) {
			fmt.Printf("Invalid note difficulty index %q\n", i)
			continue
		}
		difficultyMap[idx].Note = note
	}
	factors := make([]C.MetaFactors, len(difficultyMap))
	for i := range difficultyMap {
		difficultyMeta, ok := meta[songId][strconv.Itoa(i)][skilllength]
		if !ok {
			return fmt.Errorf("song meta not found for song %q, difficulty index %d, skill length %q", songId, i, skilllength)
		}
		factors[i] = difficultyMeta
	}
	length := songMap.Length
	effStr := make([]string, len(difficultyMap)+1)
	effStr[0] = "Efficiency"
	scoreStr := make([]string, len(difficultyMap)+1)
	scoreStr[0] = "Score Factor"
	srStr := make([]string, len(difficultyMap)+1)
	srStr[0] = "Skill Reliance"
	cells := make([][]gotgbot.RichBlockTableCell, 0, 6)
	header := make([]gotgbot.RichBlockTableCell, len(difficultyMap)+1)
	header[0] = gotgbot.RichBlockTableCell{
		Text:     gotgbot.RichTextString("Difficulty"),
		IsHeader: true,
		Align:    "center",
	}
	bpmStr := make([]string, len(difficultyMap)+1)
	noteStr := make([]string, len(difficultyMap)+1)
	bpmStr[0] = "BPM"
	noteStr[0] = "Notes"
	for i, difficulty := range difficultyMap {
		effStr[i+1] = strconv.Itoa(int(math.Round(factors[i].Efficiency(true, 1.0, 1.1, length)*100))) + "%"
		scoreStr[i+1] = strconv.Itoa(int(math.Round(factors[i].ScoreFactor(true, 1.0, 1.1)*100))) + "%"
		srStr[i+1] = strconv.Itoa(int(math.Round(factors[i].SkillReliance(true)*100))) + "%"
		header[i+1] = gotgbot.RichBlockTableCell{
			Text:     gotgbot.RichTextString(strconv.Itoa(difficulty.PlayLevel)),
			IsHeader: true,
			Align:    "center",
		}
		bpmStr[i+1] = strconv.Itoa(difficulty.Bpm)
		noteStr[i+1] = strconv.Itoa(difficulty.Note)
	}
	cells = append(cells, header)

	appendRow := func(values []string) {
		row := make([]gotgbot.RichBlockTableCell, len(values))
		for i, value := range values {
			row[i] = gotgbot.RichBlockTableCell{
				Text:  gotgbot.RichTextString(value),
				Align: "center",
			}
		}
		cells = append(cells, row)
	}

	appendRow(bpmStr)
	appendRow(noteStr)
	appendRow(effStr)
	appendRow(scoreStr)
	appendRow(srStr)
	jacketURL, _, err := songs.GetJacket(songId, songMap.JacketImage[0], "jp")
	if err != nil {
		log.Printf("Error occurred while fetching jacket URL: %v", err)
		return err
	}
	richMessage := gotgbot.InputRichMessage{
		Blocks: []gotgbot.InputRichBlock{
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(fmt.Sprintf("Song Meta for %s", songMap.MusicTitle[0])),
				Size: 3,
			},
			gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByURL(jacketURL),
				},
			},
			gotgbot.InputRichBlockTable{
				Cells: cells,
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
