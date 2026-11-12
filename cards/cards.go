package cards

import (
	"bytes"
	"context"
	"encoding/json/v2"
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
	"time"

	"github.com/gogpu/gg"
	"github.com/gogpu/gg/text"
	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	"golang.org/x/image/draw"
)

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

func retrieveCardsPath(regionCode, resourceSetName, cardStat string) string {
	url := "https://bestdori.com/assets/" + regionCode + "/characters/resourceset/" + resourceSetName + "_rip/card_" + cardStat + ".png"
	return url
}

func GetCard(cardId string) (string, string, error) {
	cardMap := utils.ReadCards()
	card, exists := cardMap[cardId]
	if !exists {
		return "", "", fmt.Errorf("card %s not found", cardId)
	}
	regionCode := regionCodeFromCard(card, "jp")
	normalPath := retrieveCardsPath(regionCode, card.ResourceSetName, "normal")
	trainingPath := retrieveCardsPath(regionCode, card.ResourceSetName, "after_training")
	normalValid := utils.IsValidPNGUrl(normalPath)
	trainingValid := utils.IsValidPNGUrl(trainingPath)
	if !normalValid {
		normalPath = ""
	}
	if !trainingValid {
		trainingPath = ""
	}
	if !normalValid && !trainingValid {
		return "", "", fmt.Errorf("no valid card image found for card ID %s", cardId)
	}
	return normalPath, trainingPath, nil
}

func GetDetailedCard(cardId string) (C.CardDetailed, error) {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(context, http.MethodGet, "https://bestdori.com/api/cards/"+cardId+".json", nil)
	if err != nil {
		return C.CardDetailed{}, err
	}
	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return C.CardDetailed{}, err
	}
	defer resp.Body.Close()
	var cardDetailed C.CardDetailed
	err = json.UnmarshalRead(resp.Body, &cardDetailed)
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
	if utils.IsValidPNGUrl(normalPath) {
		return normalPath, nil
	}
	if utils.IsValidPNGUrl(afterTrainingPath) {
		return afterTrainingPath, nil
	}
	return "", fmt.Errorf("no valid thumbnail found for card ID %s", cardId)
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

func drawScaledImage(dc *gg.Context, imgBuf image.Image, targetX, targetY, targetW, targetH float64) {
	image := scaleImagePNG(imgBuf, int(targetW), int(targetH))
	img := gg.ImageBufFromImage(image)
	dc.DrawImage(img, targetX, targetY)
}

func scaleImagePNG(src image.Image, targetWidth, targetHeight int) image.Image {
	bounds := src.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()

	ratio := float64(targetWidth) / float64(srcW)
	if hRatio := float64(targetHeight) / float64(srcH); hRatio < ratio {
		ratio = hRatio
	}
	dstW, dstH := int(float64(srcW)*ratio), int(float64(srcH)*ratio)
	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}

	dstRect := image.Rect(0, 0, dstW, dstH)
	dst := image.NewRGBA(dstRect)
	draw.CatmullRom.Scale(dst, dstRect, src, bounds, draw.Src, nil)

	return dst
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
	context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
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
	defer dc.Close()
	dc.ClearWithColor(gg.White)
	reqThumb, err := http.NewRequestWithContext(context, http.MethodGet, thumbPath, nil)
	if err != nil {
		return nil, err
	}
	respThumb, err := utils.HttpClient.Do(reqThumb)
	if err != nil {
		return nil, err
	}
	defer respThumb.Body.Close()
	baseImg, err := png.Decode(respThumb.Body)
	if err != nil {
		return nil, err
	}
	drawScaledImage(dc, baseImg, 0, 0, 180, 180)
	reqFrame, err := http.NewRequestWithContext(context, http.MethodGet, frame, nil)
	if err != nil {
		return nil, err
	}
	respFrame, err := utils.HttpClient.Do(reqFrame)
	if err != nil {
		return nil, err
	}
	defer respFrame.Body.Close()
	frameImg, err := png.Decode(respFrame.Body)
	if err != nil {
		return nil, err
	}
	drawScaledImage(dc, frameImg, 0, 0, 180, 180)
	reqAttrFrame, err := http.NewRequestWithContext(context, http.MethodGet, attrFrame, nil)
	if err != nil {
		return nil, err
	}
	respAttrFrame, err := utils.HttpClient.Do(reqAttrFrame)
	if err != nil {
		return nil, err
	}
	defer respAttrFrame.Body.Close()
	attrFrameImg, err := loadSVG(46, 46, respAttrFrame.Body)
	if err != nil {
		return nil, err
	}
	drawScaledImage(dc, attrFrameImg, 132, 2, 46, 46)
	reqStar, err := http.NewRequestWithContext(context, http.MethodGet, starImg, nil)
	if err != nil {
		return nil, err
	}
	respStar, err := utils.HttpClient.Do(reqStar)
	if err != nil {
		return nil, err
	}
	defer respStar.Body.Close()
	starImgDecoded, err := png.Decode(respStar.Body)
	if err != nil {
		return nil, err
	}
	for i := range card.Rarity {
		drawScaledImage(dc, starImgDecoded, 8, float64(147-i*25), 26, 26)
	}
	if isValidRemoteAssetURL(bandFrame) {
		reqBandFrame, err := http.NewRequestWithContext(context, http.MethodGet, bandFrame, nil)
		if err != nil {
			return nil, err
		}
		respBandFrame, err := utils.HttpClient.Do(reqBandFrame)
		if err != nil {
			return nil, err
		}
		defer respBandFrame.Body.Close()
		bandFrameImg, err := loadSVG(50, 50, respBandFrame.Body)
		if err != nil {
			return nil, err
		}
		drawScaledImage(dc, bandFrameImg, 0, 0, 50, 50)
	}
	font, err := utils.GetFontsFile("NotoSans-Regular.ttf")
	if err != nil {
		return nil, err
	}
	fontSrc, err := text.NewFontSource(font)
	dc.SetFont(fontSrc.Face(32))
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
