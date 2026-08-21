package cards

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	"golang.org/x/image/font"
)

func cardOrdinaryType(types string) bool {
	return types == "permanent" || types == "limited" || types == "dreamfes" || types == "event"
}

func regionCodeFromCard(card C.Card, preferCode string) string {
	index := 0
	switch preferCode {
	case "jp":
		index = 0
	case "en":
		index = 1
	case "tw":
		index = 2
	case "cn":
		index = 3
	case "kr":
		index = 4
	}
	if index < len(card.ReleasedAt) && card.ReleasedAt[index] != "" && card.ReleasedAt[index] != "null" {
		return preferCode
	}
	for i, releasedAt := range card.ReleasedAt {
		if releasedAt == "" || releasedAt == "null" {
			continue
		}
		switch i {
		case 0:
			return "jp"
		case 1:
			return "en"
		case 2:
			return "tw"
		case 3:
			return "cn"
		case 4:
			return "kr"
		default:
			return "jp"
		}
	}
	return "jp"
}

// retrieveCardsPath retrieves the path for a card image based on its region, resource set name, and stat, then returns its url.
// regionCode: "en" for English, "jp" for Japanese, etc.
// resourceSetName: The resource set name of the card.
// cardStat: The card stat, e.g., "normal", "after_training".
func retrieveCardsPath(regionCode string, resourceSetName string, cardStat string) (string, error) {
	// Implementation for retrieving card image paths
	url := "https://bestdori.com/assets/" + regionCode + "/characters/resourceset/" + resourceSetName + "_rip/card_" + cardStat + ".png"
	return url, nil
}

func GetCard(cardId string) (string, string, error) {
	cardMap := utils.ReadCards()
	card, exists := cardMap[cardId]
	if !exists {
		return "", "", nil
	}
	regionCode := regionCodeFromCard(card, "jp")
	if card.Rarity < 3 || !cardOrdinaryType(card.Type) {
		if card.Type == "campaign" {
			if card.Rarity < 3 {
				normalPath, err := retrieveCardsPath(regionCode, card.ResourceSetName, "normal")
				if err != nil {
					return "", "", err
				}
				return normalPath, "", nil
			}
			trainingPath, err := retrieveCardsPath(regionCode, card.ResourceSetName, "after_training")
			if err != nil {
				return "", "", err
			}
			return "", trainingPath, nil
		}
		if card.Rarity < 3 {

			normalPath, err := retrieveCardsPath(regionCode, card.ResourceSetName, "normal")
			if err != nil {
				return "", "", err
			}
			return normalPath, "", nil
		}
		if card.Type == "birthday" || card.Type == "others" || card.Type == "kirafes" {
			trainingPath, err := retrieveCardsPath(regionCode, card.ResourceSetName, "after_training")
			if err != nil {
				return "", "", err
			}
			return "", trainingPath, nil
		}

	}
	normalPath, err := retrieveCardsPath(regionCode, card.ResourceSetName, "normal")
	if err != nil {
		return "", "", err
	}
	trainingPath, err := retrieveCardsPath(regionCode, card.ResourceSetName, "after_training")
	if err != nil {
		return "", "", err
	}
	return normalPath, trainingPath, nil
}

func GetDetailedCard(cardId string) (C.CardDetailed, error) {
	resp, err := http.Get("https://bestdori.com/api/cards/" + cardId + ".json")
	if err != nil {
		return C.CardDetailed{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return C.CardDetailed{}, err
	}
	var cardDetailed C.CardDetailed
	err = json.Unmarshal(body, &cardDetailed)
	if err != nil {
		return C.CardDetailed{}, err
	}
	return cardDetailed, nil
}

func getThumbCard(cardId string) (string, error) {
	cardMap := utils.ReadCards()
	card, exists := cardMap[cardId]
	if !exists {
		return "", nil
	}
	regionCode := regionCodeFromCard(card, "jp")
	cardIDint, _ := strconv.Atoi(cardId)
	// 每50张卡片为一组，计算出卡片所在的组号，并格式化为5位数的字符串
	cardPack := fmt.Sprintf("%05d", cardIDint/50)
	normalPath := "https://bestdori.com/assets/" + regionCode + "/thumb/chara/card" + cardPack + "_rip/" + card.ResourceSetName + "_normal.png"
	afterTrainingPath := "https://bestdori.com/assets/" + regionCode + "/thumb/chara/card" + cardPack + "_rip/" + card.ResourceSetName + "_after_training.png"
	if card.Rarity < 3 || !cardOrdinaryType(card.Type) {
		if card.Type == "campaign" {
			if card.Rarity < 3 {
				return normalPath, nil
			}
			return afterTrainingPath, nil
		}
		if card.Rarity < 3 {
			return normalPath, nil
		}
		if card.Type == "birthday" || card.Type == "others" || card.Type == "kirafes" {
			return afterTrainingPath, nil
		}
	}
	return normalPath, nil
}

func loadSVG(w, h float64, stream io.Reader) (image.Image, error) {
	svgData, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	svgStr, err := cleanSVGXML(string(svgData))
	if err != nil {
		return nil, err
	}
	icon, err := canvas.ParseSVG(strings.NewReader(svgStr))
	if err != nil {
		return nil, err
	}
	origW, origH := icon.Size()
	dpmmW := w / origW
	dpmmH := h / origH
	dpmm := math.Min(dpmmW, dpmmH)
	img := rasterizer.Draw(icon, canvas.DPMM(dpmm), canvas.DefaultColorSpace)
	return img, nil
}

func drawScaledImage(dc *gg.Context, img image.Image, targetX, targetY, targetW, targetH float64) {
	bounds := img.Bounds()
	origW := float64(bounds.Dx())
	origH := float64(bounds.Dy())

	// 动态计算缩放比例
	scaleX := targetW / origW
	scaleY := targetH / origH

	dc.Push()
	dc.Translate(targetX, targetY)
	dc.Scale(scaleX, scaleY)
	dc.DrawImage(img, 0, 0)
	dc.Pop()
}

func frameHelper(rarity int, attribute string) string {
	if rarity == -1 {
		switch attribute {
		case "powerful":
			return C.AttributeFrames[0]
		case "cool":
			return C.AttributeFrames[1]
		case "happy":
			return C.AttributeFrames[2]
		case "pure":
			return C.AttributeFrames[3]
		}
	} else {
		switch rarity {
		case 1:
			switch attribute {
			case "powerful":
				return C.Card1PowerfulFrame
			case "cool":
				return C.Card1CoolFrame
			case "happy":
				return C.Card1HappyFrame
			case "pure":
				return C.Card1PureFrame
			}
		case 2:
			return C.Card2Frame
		case 3:
			return C.Card3Frame
		case 4:
			return C.Card4Frame
		case 5:
			return C.Card5Frame
		}
	}
	return ""
}

func ThumbGenerate(cardId string, percent int) (image.Image, error) {
	cardMap := utils.ReadCards()
	card, exists := cardMap[cardId]
	if !exists {
		return nil, errors.New("card not found")
	}
	thumbPath, err := getThumbCard(cardId)
	if err != nil {
		return nil, err
	}
	starImg := C.StarIcon
	frame := frameHelper(card.Rarity, card.Attribute)
	attrFrame := frameHelper(-1, card.Attribute)
	chara := utils.ReadCharacters()
	character := chara[fmt.Sprintf("%d", card.CharacterID)]

	bandFrame := C.BandFrames[character.BandID-1]
	dc := gg.NewContext(180, 250)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	resp, err := http.Get(thumbPath)
	if err != nil {
		return nil, err
	}
	baseImg, _ := png.Decode(resp.Body)
	drawScaledImage(dc, baseImg, 0, 0, 180, 180)
	resp, err = http.Get(frame)
	if err != nil {
		return nil, err
	}
	frameImg, _ := png.Decode(resp.Body)
	drawScaledImage(dc, frameImg, 0, 0, 180, 180)
	resp, err = http.Get(attrFrame)
	if err != nil {
		return nil, err
	}
	attrFrameImg, err := loadSVG(46, 46, resp.Body)
	if err != nil {
		return nil, err
	}
	drawScaledImage(dc, attrFrameImg, 132, 2, 46, 46)
	resp, err = http.Get(starImg)
	if err != nil {
		return nil, err
	}
	starImgDecoded, _ := png.Decode(resp.Body)
	for i := 0; i < card.Rarity; i++ {
		drawScaledImage(dc, starImgDecoded, 8, float64(147-i*25), 26, 26)
	}
	if isValidRemoteAssetURL(bandFrame) {
		resp, err = http.Get(bandFrame)
		if err != nil {
			return nil, err
		}
		bandFrameImg, err := loadSVG(50, 50, resp.Body)
		if err != nil {
			return nil, err
		}
		drawScaledImage(dc, bandFrameImg, 0, 0, 50, 50)
	}
	font, err := utils.FontFS.ReadFile("fonts/NotoSansSC-Regular.ttf")
	if err != nil {
		return nil, err
	}
	fontFace, err := loadEmbedTTF(font, 32)
	if err != nil {
		return nil, err
	}
	dc.SetFontFace(fontFace)
	dc.SetHexColor("#000000")
	var inputStr string
	if percent > 0 {
		inputStr = fmt.Sprintf("%s (+%d%%)", cardId, percent)
	} else {
		inputStr = cardId
	}
	dc.DrawStringAnchored(inputStr, 90, 220, 0.5, 0.5)
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

func processNode(node XMLNode) *XMLNode {
	tag := strings.ToLower(node.XMLName.Local)

	// 过滤 oksvg 不需要或不支持的标签
	if tag == "metadata" || tag == "image" || (tag == "defs" && len(node.Children) == 0) {
		return nil
	}

	var newAttrs []xml.Attr
	isDisplayNone := false

	// 处理属性，将 style="key:val;" 转为标准的 SVG 属性
	for _, attr := range node.Attrs {
		if attr.Name.Local == "style" {
			styles := strings.SplitSeq(attr.Value, ";")
			for s := range styles {
				s = strings.TrimSpace(s)
				if s == "" {
					continue
				}
				kv := strings.SplitN(s, ":", 2)
				if len(kv) == 2 {
					k := strings.TrimSpace(kv[0])
					v := strings.TrimSpace(kv[1])

					if k == "display" && v == "none" {
						isDisplayNone = true
					} else if k != "display" {
						newAttrs = append(newAttrs, xml.Attr{
							Name:  xml.Name{Local: k},
							Value: v,
						})
					}
				}
			}
		} else if attr.Name.Local == "display" && attr.Value == "none" {
			isDisplayNone = true
		} else {
			newAttrs = append(newAttrs, attr)
		}
	}

	// 如果当前图层设置了 display:none，直接丢弃该节点
	if isDisplayNone {
		return nil
	}

	node.Attrs = newAttrs

	// 递归处理子节点
	var newChildren []XMLNode
	for _, child := range node.Children {
		if processed := processNode(child); processed != nil {
			newChildren = append(newChildren, *processed)
		}
	}
	node.Children = newChildren

	return &node
}

func isValidRemoteAssetURL(raw string) bool {
	if raw == "" || raw == "0" || raw == "null" {
		return false
	}
	return strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://")
}

func cleanSVGXML(input string) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(input))
	var root XMLNode
	if err := decoder.Decode(&root); err != nil {
		return "", err
	}

	cleanedRoot := processNode(root)
	if cleanedRoot == nil {
		return "", fmt.Errorf("清洗后 SVG 内容为空")
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(cleanedRoot); err != nil {
		return "", err
	}

	return buf.String(), nil
}

/*
func GetGachaVoice(resourceSetName string, cardType string, regionCode string) (string, error) {
	switch cardType {
	case "birthday":
		cardType = "birthday"
	case "permanent":
		cardType = "operation"
	case "limited", "dreamfes":
		cardType = "limited"
	}
	url := "https://bestdori.com/assets/" + regionCode + "/gacha/voice/" + cardType + "/spin_rip/" + resourceSetName + ".mp3"
	return url, nil
}
*/
