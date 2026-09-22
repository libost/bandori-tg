package songs

import (
	"encoding/json"
	"fmt"
	"image"
	"net/http"
	"strconv"

	"golang.org/x/image/font"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
)

func GetSongs(songId string) (C.Songs, error) {
	songsMap := utils.ReadSongs()
	song, exists := songsMap[songId]
	if !exists {
		return C.Songs{}, nil
	}
	return song, nil
}

func GetDetailedSong(songId string) (C.SongsDetailed, error) {
	songsMap := utils.ReadSongs()
	_, exists := songsMap[songId]
	if !exists {
		return C.SongsDetailed{}, nil
	}
	url := "https://bestdori.com/api/songs/" + songId + ".json"
	resp, err := http.Get(url)
	if err != nil {
		return C.SongsDetailed{}, err
	}
	defer resp.Body.Close()
	songDetails := json.NewDecoder(resp.Body)
	var detailedSong C.SongsDetailed
	err = songDetails.Decode(&detailedSong)
	if err != nil {
		return C.SongsDetailed{}, err
	}
	return detailedSong, nil
}

func ceilTo10(n int) int {
	return (n + 9) / 10 * 10
}

func getJacket(songId, musicJacket, region string) (string, string, error) {
	songIdInt, err := strconv.Atoi(songId)
	if err != nil {
		return "", "", err
	}
	musicJacketPack := ceilTo10(songIdInt)
	jacketURL := fmt.Sprintf("https://bestdori.com/assets/%s/musicjacket/musicjacket%d_rip/assets-star-forassetbundle-startapp-musicjacket-musicjacket%d-%s-jacket.png", region, musicJacketPack, musicJacketPack, musicJacket)
	thumbURL := fmt.Sprintf("https://bestdori.com/assets/%s/musicjacket/musicjacket%d_rip/assets-star-forassetbundle-startapp-musicjacket-musicjacket%d-%s-thumb.png", region, musicJacketPack, musicJacketPack, musicJacket)
	return jacketURL, thumbURL, nil
}

func getBgm(songId, region string) (string, error) {
	songIdInt, err := strconv.Atoi(songId)
	if err != nil {
		return "", err
	}
	songId = fmt.Sprintf("%03d", songIdInt)
	url := fmt.Sprintf("https://bestdori.com/assets/%s/sound/bgm%s_rip/bgm%s.mp3", region, songId, songId)
	return url, nil
}

type difficultyData struct {
	level []int
}

// generateDifficultyImg generates an image representing the difficulty levels of a song.
// We may use this function later.
// WARNING: THIS FUNCTION IS NOT TESTED. A TEST IS REQUIRED BEFORE USING IT.
func generateDifficultyImg(diffData difficultyData) (image.Image, error) {
	dc := gg.NewContext(len(diffData.level)*20, 20)
	dc.SetHexColor("#FFFFFF")
	dc.Clear()
	for i, level := range diffData.level {
		dc.DrawCircle(float64(10+i*20), 10, 20)
		dc.SetHexColor(C.DifficultyColors[i])
		dc.Fill()
		font, err := utils.GetFontsFile("NotoSans-Regular.ttf")
		if err != nil {
			return nil, err
		}
		ttfont, err := loadEmbedTTF(font, 12)
		if err != nil {
			return nil, err
		}
		dc.SetFontFace(ttfont)
		dc.SetHexColor("#000000")
		inputStr := strconv.Itoa(level)
		dc.DrawStringAnchored(inputStr, float64(10+i*20), 10, 0.5, 0.5)
	}
	return dc.Image(), nil
}

func loadEmbedTTF(data []byte, points float64) (font.Face, error) {
	f, err := truetype.Parse(data)
	if err != nil {
		return nil, err
	}
	face := truetype.NewFace(f, &truetype.Options{
		Size: points,
		DPI:  72,
	})
	return face, nil
}
