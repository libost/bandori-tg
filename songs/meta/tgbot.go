package meta

import (
	"fmt"
	"log"
	"math"
	"strconv"

	C "github.com/libost/bandori-tg/constants"
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
	meta := utils.ReadSongsMeta()
	songId := "1"      // Temporary hardcoded song ID for demonstration purposes
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
	efficiencies := make([]int, len(difficultyMap))
	for i := range difficultyMap {
		efficiencies[i] = int(math.Round(factors[i].Efficiency(true, 1.0, 1.1, length) * 100))
	}
	effStr := make([]string, len(efficiencies)+1)
	effStr[0] = "Efficiency"
	for i, eff := range efficiencies {
		effStr[i+1] = strconv.Itoa(eff) + "%"
	}
	scores := make([]int, len(difficultyMap))
	for i := range difficultyMap {
		scores[i] = int(math.Round(factors[i].ScoreFactor(true, 1.0, 1.1) * 100))
	}
	scoreStr := make([]string, len(scores)+1)
	scoreStr[0] = "Score Factor"
	for i, score := range scores {
		scoreStr[i+1] = strconv.Itoa(score) + "%"
	}
	srs := make([]int, len(difficultyMap))
	for i := range difficultyMap {
		srs[i] = int(math.Round(factors[i].SkillReliance(true) * 100))
	}
	srStr := make([]string, len(srs)+1)
	srStr[0] = "Skill Reliance"
	for i, sr := range srs {
		srStr[i+1] = strconv.Itoa(sr) + "%"
	}
	cells := make([][]gotgbot.RichBlockTableCell, 0, 6)
	header := make([]gotgbot.RichBlockTableCell, len(difficultyMap)+1)
	header[0] = gotgbot.RichBlockTableCell{
		Text:     gotgbot.RichTextString("Difficulty"),
		IsHeader: true,
		Align:    "center",
	}
	for i, difficulty := range difficultyMap {
		header[i+1] = gotgbot.RichBlockTableCell{
			Text:     gotgbot.RichTextString(strconv.Itoa(difficulty.PlayLevel)),
			IsHeader: true,
			Align:    "center",
		}
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

	bpms := make([]int, len(difficultyMap))
	notes := make([]int, len(difficultyMap))
	for i, difficulty := range difficultyMap {
		bpms[i] = difficulty.Bpm
		notes[i] = difficulty.Note
	}
	bpmStr := make([]string, len(bpms)+1)
	noteStr := make([]string, len(notes)+1)
	bpmStr[0] = "BPM"
	noteStr[0] = "Notes"
	for i, bpm := range bpms {
		bpmStr[i+1] = strconv.Itoa(bpm)
	}
	for i, note := range notes {
		noteStr[i+1] = strconv.Itoa(note)
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
