package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"github.com/libost/bandori-tg/cards"
	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
	"golang.org/x/image/font/opentype"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func GetDetailedEvents(eventID string) (C.EventDetailed, error) {
	resp, err := http.Get("https://bestdori.com/api/events/" + eventID + ".json")
	if err != nil {
		return C.EventDetailed{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return C.EventDetailed{}, err
	}
	var event C.EventDetailed
	err = json.Unmarshal([]byte(data), &event)
	if err != nil {
		return C.EventDetailed{}, err
	}
	return event, nil
}

func regionCodeFromEvent(event C.EventDetailed, preferCode string) string {
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
	if index < len(event.EndAt) && event.EndAt[index] != "" && event.EndAt[index] != "null" {
		return preferCode
	}
	for i, releasedAt := range event.EndAt {
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

func generateEventMembers(event C.EventDetailed) (image.Image, error) {
	members := event.Members
	dc := gg.NewContext(len(members)*180+20*len(members)+20, 300)
	dc.SetHexColor("#FFFFFF")
	dc.Clear()
	for i, member := range members {
		memberImage, err := cards.ThumbGenerate(fmt.Sprintf("%d", member.SituationID), member.Percent)
		if err != nil {
			return nil, err
		}
		dc.DrawImageAnchored(memberImage, 200*i+20, 25, 0, 0)
	}
	return dc.Image(), nil
}

func generateEventBonusCards(event C.EventDetailed) (image.Image, error) {
	bonusCards := event.RewardCards
	dc := gg.NewContext(len(bonusCards)*180+20*len(bonusCards)+20, 300)
	dc.SetHexColor("#FFFFFF")
	dc.Clear()
	for i, card := range bonusCards {
		cardImage, err := cards.ThumbGenerate(fmt.Sprintf("%d", card), 0)
		if err != nil {
			return nil, err
		}
		dc.DrawImageAnchored(cardImage, 200*i+20, 25, 0, 0)
	}
	return dc.Image(), nil
}

func setFont(fontFile string) (*opentype.Font, error) {
	file, err := utils.GetFontsFile(fontFile)
	if err != nil {
		return nil, err
	}
	ttFont, err := opentype.Parse(file)
	if err != nil {
		return nil, err
	}
	fontName := fontFile
	font.DefaultCache.Add([]font.Face{
		{
			Font: font.Font{Typeface: font.Typeface(fontName)},
			Face: ttFont,
		},
	})
	plot.DefaultFont = font.Font{Typeface: font.Typeface(fontName)}
	plotter.DefaultFont = plot.DefaultFont
	return ttFont, nil
}

func getEventTracker(regionCode, eventID string, tier int) (image.Image, int64, int64, int64, int64, error) {
	eventMap := utils.ReadEvents()
	event, exists := eventMap[eventID]
	if !exists {
		return nil, 0, 0, 0, 0, C.ErrNoSuchEvent
	}
	var tierList []int
	switch regionCode {
	case "jp":
		tierList = C.JPTierList[:]
	case "tw":
		tierList = C.TWTierList[:]
	case "cn":
		tierList = C.CNTierList[:]
	case "kr":
		tierList = C.KRTierList[:]
	case "en":
		tierList = C.ENTierList[:]
	}
	var region int
	if !slices.Contains(tierList, tier) {
		region, tier = getRegionCodeandTier(regionCode)
	} else {
		region, _ = getRegionCodeandTier(regionCode)
	}

	url := fmt.Sprintf("https://bestdori.com/api/tracker/data?server=%d&event=%s&tier=%d", region, eventID, tier)
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	var tracker C.EventTracker
	err = json.Unmarshal(data, &tracker)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	if len(tracker.Cutoffs) == 0 {
		for _, tier := range tierList {
			// Process each tier
			url := fmt.Sprintf("https://bestdori.com/api/tracker/data?server=%d&event=%s&tier=%d", region, eventID, tier)
			resp, err := http.Get(url)
			if err != nil {
				continue
			}
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}
			var newTracker C.EventTracker
			err = json.Unmarshal(data, &newTracker)
			if err != nil {
				continue
			}
			if len(newTracker.Cutoffs) > 0 {
				tracker = newTracker
				break
			}
		}
		if len(tracker.Cutoffs) == 0 {
			return nil, 0, 0, 0, 0, C.ErrNoCutoffData
		}
	}
	startAt, err := strconv.ParseInt(event.StartAt[region], 10, 64)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	endAt, err := strconv.ParseInt(event.EndAt[region], 10, 64)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	predicted, err := getPredictedEp(event.EventType, region, tier, tracker, startAt, endAt)
	if err != nil && !errors.Is(err, C.ErrCannotPredict) {
		return nil, 0, 0, 0, 0, err
	}
	var lastPredictedEp int64
	if len(predicted) > 0 {
		lastPredictedEp = predicted[len(predicted)-1].Ep
	} else {
		lastPredictedEp = 0
	}
	var fontFile string
	switch regionCode {
	case "jp":
		fontFile = "NotoSansJP-Regular.ttf"
	case "tw":
		fontFile = "NotoSansTC-Regular.ttf"
	case "cn":
		fontFile = "NotoSansSC-Regular.ttf"
	case "kr":
		fontFile = "NotoSansKR-Regular.ttf"
	case "en":
		fontFile = "NotoSans-Regular.ttf"
	}
	_, err = setFont(fontFile)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	plot.DefaultFont = font.Font{Typeface: font.Typeface(fontFile)}
	plotter.DefaultFont = plot.DefaultFont

	pts := make(plotter.XYs, len(tracker.Cutoffs))
	predictedPts := make(plotter.XYs, len(predicted))
	phonyPts := make(plotter.XYs, 1)

	for i, cutoff := range tracker.Cutoffs {
		pts[i].X = float64(cutoff.Time / 1000)
		pts[i].Y = float64(cutoff.Ep)
	}
	for i, cutoff := range predicted {
		predictedPts[i].X = float64(cutoff.Time / 1000)
		predictedPts[i].Y = float64(cutoff.Ep)
	}
	phonyPtsY := float64(pts[len(pts)-1].Y-pts[0].Y) * 1.3
	if len(predictedPts) > 0 {
		phonyPtsY = float64(predictedPts[len(predictedPts)-1].Y-pts[0].Y) * 1.3
	}
	phonyPts[0].X = float64(endAt / 1000)
	phonyPts[0].Y = phonyPtsY
	p := plot.New()
	p.Title.Text = fmt.Sprintf("%s(%s) - %s - T%d", event.EventName[region], eventID, strings.ToUpper(regionCode), tier)
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Points"
	p.X.Tick.Marker = plot.TimeTicks{
		Format: "01-02", // 格式化时间字符串，例如 "2006-01-02" 或 "15:04"
	}
	p.Y.Tick.Marker = customTicker{}
	grid := plotter.NewGrid()
	grid.Vertical.Color = color.RGBA{R: 230, G: 230, B: 230, A: 255}
	grid.Horizontal.Color = color.RGBA{R: 230, G: 230, B: 230, A: 255}
	p.Add(grid)
	line, points, err := plotter.NewLinePoints(pts)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	line.Color = color.RGBA{R: 31, G: 119, B: 180, A: 255} // 经典蓝色
	line.Width = vg.Points(2)

	// 数据节点样式（实心圆点）
	points.Shape = draw.CircleGlyph{}
	points.Color = color.RGBA{R: 31, G: 119, B: 180, A: 255}
	points.Radius = vg.Points(3)
	p.Add(line, points)

	line2, points2, err := plotter.NewLinePoints(predictedPts)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	line2.Color = color.RGBA{R: 255, G: 127, B: 14, A: 255} // 经典橙色
	line2.Width = vg.Points(2)

	// 预测数据节点样式（空心圆点）
	points2.Shape = draw.CircleGlyph{}
	points2.Color = color.RGBA{R: 255, G: 127, B: 14, A: 255}
	points2.Radius = vg.Points(3)
	p.Add(line2, points2)

	// 添加一个透明的点和线，用于把图片拉长到活动结束时间
	line3, point3, err := plotter.NewLinePoints(phonyPts)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	line3.Color = color.RGBA{R: 255, G: 255, B: 255, A: 0}
	line3.Width = vg.Points(2)
	point3.Shape = draw.CircleGlyph{}
	point3.Color = color.RGBA{R: 255, G: 255, B: 255, A: 0}
	point3.Radius = vg.Points(3)
	p.Add(line3, point3)

	p.Legend.Add("Actual Cutoff", line, points)
	p.Legend.Add("Predicted Final Cutoff", line2, points2)

	canvas := vgimg.New(10*vg.Inch, 4*vg.Inch)
	dc := draw.New(canvas)
	p.Draw(dc)
	lastEpIncrement := tracker.Cutoffs[len(tracker.Cutoffs)-1].Ep - tracker.Cutoffs[len(tracker.Cutoffs)-2].Ep
	lastTimeIncrement := tracker.Cutoffs[len(tracker.Cutoffs)-1].Time - tracker.Cutoffs[len(tracker.Cutoffs)-2].Time
	speed := int64(float64(lastEpIncrement) / (float64(lastTimeIncrement) / 3600000)) // 计算每小时的速度
	return canvas.Image(), tracker.Cutoffs[len(tracker.Cutoffs)-1].Ep, lastPredictedEp, tracker.Cutoffs[len(tracker.Cutoffs)-1].Time, speed, nil
}

type predictedCutoff struct {
	Time    int64
	Ep      int64
	Percent float64
}

func getPredictedEp(eventType string, server, tier int, tracker C.EventTracker, timeStart, timeEnd int64) ([]predictedCutoff, error) {
	latestTime := tracker.Cutoffs[len(tracker.Cutoffs)-1].Time
	predicted := make([]predictedCutoff, len(tracker.Cutoffs))
	if latestTime-timeStart < 86400000 { // 如果最新时间距离开始时间不足一天，则不进行预测
		return []predictedCutoff{}, C.ErrCannotPredict
	}
	ratesData := utils.ReadRates()
	var rateAssigned float64
	for _, rate := range ratesData {
		if rate.Type == eventType && rate.Server == server && rate.Tier == tier {
			rateAssigned = rate.Rate
			break
		}
	}
	var avgPercent, sumPercent float64
	var avgEp, sumEp int64
	for i, cutoff := range tracker.Cutoffs {
		percent := float64(cutoff.Time-timeStart) / float64(timeEnd-timeStart)
		predicted[i].Percent = percent
		sumPercent += predicted[i].Percent
		sumEp += tracker.Cutoffs[i].Ep
	}
	if len(tracker.Cutoffs) > 0 {
		avgPercent = sumPercent / float64(len(tracker.Cutoffs))
		avgEp = sumEp / int64(len(tracker.Cutoffs))
	}
	for i, cutoff := range tracker.Cutoffs {
		predicted[i].Time = cutoff.Time
		if cutoff.Time-timeStart < 86400000 { // 如果最新时间距离开始时间不足一天，则不进行预测
			predicted[i].Ep = cutoff.Ep
			continue
		}
		var z, w float64
		cutoffAltered := tracker.Cutoffs[:i+1]
		for k, cutoff := range cutoffAltered {
			z += (predicted[k].Percent - avgPercent) * (float64(cutoff.Ep) - float64(avgEp))
			w += (predicted[k].Percent - avgPercent) * (predicted[k].Percent - avgPercent)
		}
		b := z / w
		a := float64(avgEp) - b*avgPercent
		predictedEp := a + b*(float64(1)+rateAssigned) // 预测的最终ep值
		predicted[i].Ep = int64(predictedEp)
	}
	var lastEqualIndex int
	for i, cutoff := range tracker.Cutoffs {
		if cutoff.Ep != predicted[i].Ep {
			lastEqualIndex = i
			break // 如果预测值与实际值不同，则停止截断
		}
	}
	predicted = predicted[lastEqualIndex:] // 截断预测值，去掉与实际值相同的部分
	predicted = append(
		predicted,
		predictedCutoff{
			Time:    timeEnd,
			Ep:      predicted[len(predicted)-1].Ep,
			Percent: 1.0,
		},
	)
	return predicted, nil
}

func getRegionCodeandTier(regionCode string) (int, int) {
	switch regionCode {
	case "jp":
		return 0, 1000
	case "en":
		return 1, 3000
	case "tw":
		return 2, 100
	case "cn":
		return 3, 2000
	default:
		return 0, 1000
	}
}

type customTicker struct {
	plot.DefaultTicks
}

func (t customTicker) Ticks(min, max float64) []plot.Tick {
	ticks := t.DefaultTicks.Ticks(min, max)
	for i := range ticks {
		if ticks[i].Label == "" {
			continue
		}
		// 格式化大数字
		ticks[i].Label = formatCompactNumber(ticks[i].Value)
	}
	return ticks
}

func formatCompactNumber(v float64) string {
	if v == 0 {
		return "0"
	}
	abs := math.Abs(v)
	var val float64
	var suffix string

	switch {
	case abs >= 1e9:
		val, suffix = v/1e9, "B"
	case abs >= 1e6:
		val, suffix = v/1e6, "M"
	case abs >= 1e3:
		val, suffix = v/1e3, "K"
	default:
		return fmt.Sprintf("%.0f", v)
	}

	// 保留最多 1 位小数，如果是整数则去掉 .0 (例如 10.0M -> 10M, 1.5M -> 1.5M)
	str := fmt.Sprintf("%.1f", val)
	str = strings.TrimSuffix(str, ".0")
	return str + suffix
}
