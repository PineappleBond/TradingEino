package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/logger"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/internal/utils/xmd"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	"github.com/PineappleBond/TradingEino/backend/pkg/okex/models/market"
	requests_market "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/markcheno/go-talib"
	"github.com/shopspring/decimal"
	"golang.org/x/time/rate"
)

type OkxCandlesticksTool struct {
	svcCtx  *svc.ServiceContext
	limiter *rate.Limiter
}

func (c *OkxCandlesticksTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:  "okx-candlesticks-tool",
		Desc:  "调用OKX接口获取K线数据的工具",
		Extra: map[string]any{},
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"symbol": &schema.ParameterInfo{
				Type:     schema.String,
				Desc:     "交易对,比如ETH-USDT-SWAP,BTC-USDT",
				Enum:     nil,
				Required: true,
			},
			"bar": &schema.ParameterInfo{
				Type: schema.String,
				Desc: "交易周期",
				Enum: []string{
					okex.Bar1m.String(),
					okex.Bar3m.String(),
					okex.Bar5m.String(),
					okex.Bar15m.String(),
					okex.Bar30m.String(),
					okex.Bar1H.String(),
					okex.Bar2H.String(),
					okex.Bar4H.String(),
					okex.Bar6H.String(),
					okex.Bar8H.String(),
					okex.Bar12H.String(),
					okex.Bar1D.String(),
					okex.Bar1W.String(),
					okex.Bar1M.String(),
					okex.Bar3M.String(),
					okex.Bar6M.String(),
					okex.Bar1Y.String(),
				},
				Required: true,
			},
			"limit": &schema.ParameterInfo{
				Type:     schema.Number,
				Desc:     "获取的最新的N条",
				Required: true,
			},
		}),
	}, nil
}

func (c *OkxCandlesticksTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	type Request struct {
		Symbol string `json:"symbol"`
		Limit  int    `json:"limit"`
		Bar    string `json:"bar"`
	}
	var request Request
	err := json.Unmarshal([]byte(argumentsInJSON), &request)
	if err != nil {
		return "", err
	}
	candlesticks, err := c.GetCandlesticks(ctx, request.Symbol, okex.BarSize(request.Bar), nil, request.Limit+300)
	if err != nil {
		return "", err
	}
	if len(candlesticks) <= 300 {
		return "获取K线数据失败", nil
	}
	// 获取指标
	indicatorCalculator := NewIndicatorCalculator(candlesticks)
	technicalIndicators := indicatorCalculator.Calculate(request.Limit)
	rows := make([][]string, 0)
	for _, technicalIndicator := range technicalIndicators {
		rows = append(rows, technicalIndicator.Row())
	}
	if len(rows) == 0 {
		return "获取K线数据失败", nil
	}
	output := ""
	table := xmd.CreateMarkdownTable(TechnicalIndicatorsHeaders, rows)
	output += fmt.Sprintf("# 近%d根`%s`周期的K线数据", len(rows), request.Bar)
	output += "\n```markdown\n"
	output += table
	output += "\n```\n---\n"

	// 输出筹码分布
	volumeProfile := indicatorCalculator.CalculateVolumeProfile(30, 0.72, 180)
	if volumeProfile != nil {
		output += fmt.Sprintf("# 近 %d根`%s`周期的K线(%s ～ %s)数据的筹码分布", volumeProfile.KLineCount, request.Bar, volumeProfile.FirstCandlestick.Time.Format("2006年01月02日15时04分"), volumeProfile.LatestCandlestick.Time.Format("2006年01月02日15时04分"))
		output += "\n```markdown\n"

		output += `| 价格区间 |  占比 | Tag |
`
		output += `| :------ | :--- | :--- |
`
		for j := range volumeProfile.Items {
			i := len(volumeProfile.Items) - 1 - j
			item := volumeProfile.Items[i]
			ratio := item.Ratio * 100
			tag := ""
			if i == volumeProfile.PocIndex {
				tag += " [VPOC] "
			}
			if item.PriceLow <= volumeProfile.ValHigh && volumeProfile.ValHigh <= item.PriceHigh {
				tag += " [VAH] "
			}
			if item.PriceLow <= volumeProfile.ValLow && volumeProfile.ValLow <= item.PriceHigh {
				tag += " [VAL] "
			}
			output += fmt.Sprintf(
				`| %.2f - %.2f | %.2f%% | %s |
`,
				item.PriceLow, item.PriceHigh, ratio, tag,
			)
		}

		output += "\n```\n---\n"

	}

	return output, nil
}

func NewOkxCandlesticksTool(svcCtx *svc.ServiceContext) *OkxCandlesticksTool {
	return &OkxCandlesticksTool{
		svcCtx:  svcCtx,
		limiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 2), // 10 req/s for Market endpoint (burst=2)
	}
}

func (c *OkxCandlesticksTool) GetCandlesticks(
	ctx context.Context, symbol string,
	bar okex.BarSize, afterDatetime *time.Time, limit int,
) ([]*market.Candle, error) {
	candles := make([]*market.Candle, 0)
	var after int64
	if afterDatetime != nil {
		after = afterDatetime.UnixMilli()
	}
	for {
		// 等待速率限制
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, err
		}
		getCandlesticksHistory, err := c.svcCtx.OKXClient.Rest.Market.GetCandlesticksHistory(requests_market.GetCandlesticks{
			InstID: symbol,
			After:  after,
			Limit:  300,
			Bar:    bar,
		})
		if err != nil {
			return nil, err
		}
		// 倒序
		candlesTmp := getCandlesticksHistory.Candles
		sort.Slice(candlesTmp, func(i, j int) bool {
			iTime := time.Time(candlesTmp[i].TS)
			jTime := time.Time(candlesTmp[j].TS)
			return iTime.After(jTime)
		})
		if len(candlesTmp) == 0 {
			break
		}
		t1 := time.Time(candlesTmp[len(candlesTmp)-1].TS)
		after = t1.UnixMilli()
		candles = append(candles, candlesTmp...)
		if len(candles) >= limit {
			break
		}
	}

	// 按正序排列
	sort.Slice(candles, func(i, j int) bool {
		return time.Time(candles[i].TS).Before(time.Time(candles[j].TS))
	})

	if len(candles) > limit {
		candles = candles[len(candles)-limit:]
	}

	return candles, nil
}

type Candlestick struct {
	Close    *decimal.Decimal
	Open     *decimal.Decimal
	Low      *decimal.Decimal
	High     *decimal.Decimal
	Volume   *decimal.Decimal // 成交量
	Turnover *decimal.Decimal // 成交额

	Timestamp int64
	Time      time.Time
	TimeStr   string
}

// IndicatorCalculator 指标计算器
type IndicatorCalculator struct {
	candles []*Candlestick

	// 预处理的float64数组，供talib使用
	closes    []float64
	opens     []float64
	highs     []float64
	lows      []float64
	volumes   []float64
	turnovers []float64
}

// NewIndicatorCalculator 创建指标计算器
func NewIndicatorCalculator(marketCandles []*market.Candle) *IndicatorCalculator {
	candles := make([]*Candlestick, 0)
	for _, candlestick := range marketCandles {
		c := decimal.NewFromFloat(candlestick.C)
		o := decimal.NewFromFloat(candlestick.O)
		l := decimal.NewFromFloat(candlestick.L)
		h := decimal.NewFromFloat(candlestick.H)
		v := decimal.NewFromFloat(candlestick.Vol)
		ccy := decimal.NewFromFloat(candlestick.VolCcy)
		tm := time.Time(candlestick.TS)
		candles = append(candles, &Candlestick{
			Close:     &c,
			Open:      &o,
			Low:       &l,
			High:      &h,
			Volume:    &v,
			Turnover:  &ccy,
			Timestamp: tm.Unix(),
			Time:      tm,
			TimeStr:   tm.Format("2006-01-02T15:04"),
		})
	}
	calc := &IndicatorCalculator{
		candles:   candles,
		closes:    make([]float64, len(candles)),
		opens:     make([]float64, len(candles)),
		highs:     make([]float64, len(candles)),
		lows:      make([]float64, len(candles)),
		volumes:   make([]float64, len(candles)),
		turnovers: make([]float64, len(candles)),
	}

	// 预处理数据转换
	for i, c := range candles {
		if c.Close != nil {
			calc.closes[i], _ = c.Close.Float64()
		}
		if c.Open != nil {
			calc.opens[i], _ = c.Open.Float64()
		}
		if c.High != nil {
			calc.highs[i], _ = c.High.Float64()
		}
		if c.Low != nil {
			calc.lows[i], _ = c.Low.Float64()
		}
		if c.Volume != nil {
			calc.volumes[i], _ = c.Volume.Float64()
		}
		if c.Turnover != nil {
			calc.turnovers[i] = c.Turnover.InexactFloat64()
		}
	}

	return calc
}

// Calculate 计算所有技术指标
func (calc *IndicatorCalculator) Calculate(count int) []*TechnicalIndicators {
	n := len(calc.candles)
	if n == 0 {
		logger.Debug(context.Background(), "no candles")
		return nil
	}
	if n < count {
		logger.Warn(context.Background(), "Warning: number of candles is less than the number of indicators: %d", n)
		return nil
	}

	// 计算各项指标
	macd, macdSignal, macdHist := talib.Macd(calc.closes, 12, 26, 9)
	rsi6 := talib.Rsi(calc.closes, 6)
	rsi14 := talib.Rsi(calc.closes, 14)
	rsi24 := talib.Rsi(calc.closes, 24)

	adx := talib.Adx(calc.highs, calc.lows, calc.closes, 14)
	plusDI := talib.PlusDI(calc.highs, calc.lows, calc.closes, 14)
	minusDI := talib.MinusDI(calc.highs, calc.lows, calc.closes, 14)

	mfi14 := talib.Mfi(calc.highs, calc.lows, calc.closes, calc.volumes, 14)

	cci14 := talib.Cci(calc.highs, calc.lows, calc.closes, 14)
	cci20 := talib.Cci(calc.highs, calc.lows, calc.closes, 20)

	mom10 := talib.Mom(calc.closes, 10)
	roc10 := talib.Roc(calc.closes, 10)

	ad := talib.Ad(calc.highs, calc.lows, calc.closes, calc.volumes)

	bollUpper, bollMiddle, bollLower := talib.BBands(calc.closes, 20, 2, 2, 0)

	ma5 := talib.Sma(calc.closes, 5)
	ma10 := talib.Sma(calc.closes, 10)
	ma20 := talib.Sma(calc.closes, 20)
	ma60 := talib.Sma(calc.closes, 60)
	ema12 := talib.Ema(calc.closes, 12)
	ema26 := talib.Ema(calc.closes, 26)

	atr14 := talib.Atr(calc.highs, calc.lows, calc.closes, 14)

	slowK, slowD := talib.Stoch(calc.highs, calc.lows, calc.closes, 9, 3, 0, 3, 0)

	obv := talib.Obv(calc.closes, calc.volumes)

	// 组装结果
	results := make([]*TechnicalIndicators, count)
	for i := n - count; i < n; i++ {
		closeI := decimal.NewFromFloat(calc.closes[i])
		openI := decimal.NewFromFloat(calc.opens[i])
		lowI := decimal.NewFromFloat(calc.lows[i])
		highI := decimal.NewFromFloat(calc.highs[i])
		turnoverI := decimal.NewFromFloat(calc.turnovers[i])
		results[i+count-n] = &TechnicalIndicators{
			Close:      &closeI,
			Open:       &openI,
			Low:        &lowI,
			High:       &highI,
			Volume:     calc.candles[i].Volume,
			Turnover:   &turnoverI,
			Timestamp:  calc.candles[i].Timestamp,
			Time:       time.Unix(calc.candles[i].Timestamp, 0),
			TimeStr:    time.Unix(calc.candles[i].Timestamp, 0).Format("2006-01-02 15:04:05"),
			MACD:       safeGet(macd, i),
			MACDSignal: safeGet(macdSignal, i),
			MACDHist:   safeGet(macdHist, i),
			RSI6:       safeGet(rsi6, i),
			RSI14:      safeGet(rsi14, i),
			RSI24:      safeGet(rsi24, i),
			ADX:        safeGet(adx, i),
			PlusDI:     safeGet(plusDI, i),
			MinusDI:    safeGet(minusDI, i),
			MFI14:      safeGet(mfi14, i),
			CCI14:      safeGet(cci14, i),
			CCI20:      safeGet(cci20, i),
			MOM10:      safeGet(mom10, i),
			ROC10:      safeGet(roc10, i),
			AD:         safeGet(ad, i),
			BollUpper:  safeGet(bollUpper, i),
			BollMiddle: safeGet(bollMiddle, i),
			BollLower:  safeGet(bollLower, i),
			MA5:        safeGet(ma5, i),
			MA10:       safeGet(ma10, i),
			MA20:       safeGet(ma20, i),
			MA60:       safeGet(ma60, i),
			EMA12:      safeGet(ema12, i),
			EMA26:      safeGet(ema26, i),
			ATR14:      safeGet(atr14, i),
			K:          safeGet(slowK, i),
			D:          safeGet(slowD, i),
			J:          3*safeGet(slowK, i) - 2*safeGet(slowD, i), // J = 3K - 2D
			OBV:        safeGet(obv, i),
		}
	}

	return results
}

// VolumeProfileItem 筹码分布区间
type VolumeProfileItem struct {
	PriceLow  float64 `json:"priceLow"`  // 区间下界
	PriceHigh float64 `json:"priceHigh"` // 区间上界
	Volume    float64 `json:"volume"`    // 区间内成交量
	Ratio     float64 `json:"ratio"`     // 占总成交量比例
}

// VolumeProfile 筹码分布结果
type VolumeProfile struct {
	Items             []VolumeProfileItem `json:"items"`             // 各区间筹码
	PocPrice          float64             `json:"pocPrice"`          // 控盘价格 (Point of Control)，成交量最大的区间中点
	PocIndex          int                 `json:"pocIndex"`          // 控盘区间索引
	ValLow            float64             `json:"valLow"`            // 价值区间下界 (Value Area Low)，覆盖70%成交量
	ValHigh           float64             `json:"valHigh"`           // 价值区间上界 (Value Area High)
	KLineCount        int                 `json:"kLineCount"`        // K线数量
	FirstCandlestick  *Candlestick        `json:"firstCandlestick"`  // 第一根
	LatestCandlestick *Candlestick        `json:"latestCandlestick"` // 最后一根
}

// CalculateVolumeProfile 计算筹码分布（Volume Profile）
// bins: 价格区间拆分数量
// vaRatio: Value Area 覆盖成交量比例，例如 0.7 表示 70%
// limit: 只分析最近 limit 根K线，<=0 时分析全部
// 加权分配策略：每根K线的成交量按 OHLC 加权分配到覆盖的价格区间
//   - Open/Close 各占 30% 权重，High/Low 各占 20% 权重
func (calc *IndicatorCalculator) CalculateVolumeProfile(bins int, vaRatio float64, limit int) *VolumeProfile {
	total := len(calc.candles)
	if total == 0 || bins <= 0 {
		return nil
	}

	// 计算分析范围
	start := 0
	if limit > 0 && limit < total {
		start = total - limit
	}
	n := total - start

	first := calc.candles[start]
	latest := calc.candles[total-1]

	// 找全局最高价和最低价
	globalHigh := calc.highs[start]
	globalLow := calc.lows[start]
	for i := start + 1; i < total; i++ {
		if calc.highs[i] > globalHigh {
			globalHigh = calc.highs[i]
		}
		if calc.lows[i] < globalLow {
			globalLow = calc.lows[i]
		}
	}

	priceRange := globalHigh - globalLow
	if priceRange <= 0 {
		// 所有K线价格相同，全部成交量堆到一个区间
		totalVol := 0.0
		for i := start; i < total; i++ {
			totalVol += calc.volumes[i]
		}
		return &VolumeProfile{
			Items: []VolumeProfileItem{{
				PriceLow:  globalLow,
				PriceHigh: globalHigh,
				Volume:    totalVol,
				Ratio:     1.0,
			}},
			PocPrice:          globalLow,
			PocIndex:          0,
			ValLow:            globalLow,
			ValHigh:           globalHigh,
			KLineCount:        n,
			FirstCandlestick:  first,
			LatestCandlestick: latest,
		}
	}

	binSize := priceRange / float64(bins)
	buckets := make([]float64, bins)

	// 将价格映射到区间索引
	priceToBin := func(price float64) int {
		idx := int((price - globalLow) / binSize)
		if idx >= bins {
			idx = bins - 1
		}
		if idx < 0 {
			idx = 0
		}
		return idx
	}

	// 加权分配：Open/Close 各30%，High/Low 各20%
	for i := start; i < total; i++ {
		vol := calc.volumes[i]
		if vol <= 0 {
			continue
		}

		openBin := priceToBin(calc.opens[i])
		closeBin := priceToBin(calc.closes[i])
		highBin := priceToBin(calc.highs[i])
		lowBin := priceToBin(calc.lows[i])

		buckets[openBin] += vol * 0.3
		buckets[closeBin] += vol * 0.3
		buckets[highBin] += vol * 0.2
		buckets[lowBin] += vol * 0.2
	}

	// 计算总量、找POC
	totalVolume := 0.0
	pocIndex := 0
	maxVol := 0.0
	for i, v := range buckets {
		totalVolume += v
		if v > maxVol {
			maxVol = v
			pocIndex = i
		}
	}

	// 构造结果
	items := make([]VolumeProfileItem, bins)
	for i := 0; i < bins; i++ {
		low := globalLow + float64(i)*binSize
		high := low + binSize
		ratio := 0.0
		if totalVolume > 0 {
			ratio = buckets[i] / totalVolume
		}
		items[i] = VolumeProfileItem{
			PriceLow:  low,
			PriceHigh: high,
			Volume:    buckets[i],
			Ratio:     ratio,
		}
	}

	pocPrice := globalLow + (float64(pocIndex)+0.5)*binSize

	// 计算 Value Area（覆盖约70%成交量的区间，从POC向两侧扩展）
	valLowIdx := pocIndex
	valHighIdx := pocIndex
	vaVolume := buckets[pocIndex]
	if vaRatio <= 0 || vaRatio > 1 {
		vaRatio = 0.7
	}
	vaTarget := totalVolume * vaRatio

	for vaVolume < vaTarget {
		expandLow := -1.0
		expandHigh := -1.0
		if valLowIdx > 0 {
			expandLow = buckets[valLowIdx-1]
		}
		if valHighIdx < bins-1 {
			expandHigh = buckets[valHighIdx+1]
		}

		if expandLow < 0 && expandHigh < 0 {
			break
		}

		if expandLow >= expandHigh {
			valLowIdx--
			vaVolume += buckets[valLowIdx]
		} else {
			valHighIdx++
			vaVolume += buckets[valHighIdx]
		}
	}

	valLow := globalLow + float64(valLowIdx)*binSize
	valHigh := globalLow + float64(valHighIdx+1)*binSize

	return &VolumeProfile{
		Items:             items,
		PocPrice:          pocPrice,
		PocIndex:          pocIndex,
		ValLow:            valLow,
		ValHigh:           valHigh,
		KLineCount:        n,
		FirstCandlestick:  first,
		LatestCandlestick: latest,
	}
}

// safeGet 安全获取数组值，避免越界
func safeGet(arr []float64, index int) float64 {
	if index < 0 || index >= len(arr) {
		return 0
	}
	return arr[index]
}

var (
	TechnicalIndicatorsHeaders = []string{
		"时间", "收盘价", "开盘价", "最低价", "最高价", "成交量", "成交额",
		"MACD", "MACDSignal", "MACDHist",
		"RSI14", "RSI24",
		"ADX14", "PlusDI14", "MinusDI14",
		"CCI14",
		"BollUpper", "BollMiddle", "BollLower",
		"MA5", "MA10", "MA20", "MA60", "EMA12", "EMA26",
		"ATR14",
		"K", "D", "J",
	}
)

func (t *TechnicalIndicators) Row() []string {
	res := make([]string, len(TechnicalIndicatorsHeaders))
	i := -1
	i++
	res[i] = t.TimeStr
	i++
	res[i] = t.Close.StringFixed(4)
	i++
	res[i] = t.Open.StringFixed(4)
	i++
	res[i] = t.Low.StringFixed(4)
	i++
	res[i] = t.High.StringFixed(4)
	i++
	res[i] = t.Volume.StringFixed(4)
	i++
	res[i] = t.Turnover.StringFixed(4)

	i++
	res[i] = fmt.Sprintf("%.4f", t.MACD)
	i++
	res[i] = fmt.Sprintf("%.4f", t.MACDSignal)
	i++
	res[i] = fmt.Sprintf("%.4f", t.MACDHist)

	i++
	res[i] = fmt.Sprintf("%.4f", t.RSI14)
	i++
	res[i] = fmt.Sprintf("%.4f", t.RSI24)

	i++
	res[i] = fmt.Sprintf("%.4f", t.ADX)
	i++
	res[i] = fmt.Sprintf("%.4f", t.PlusDI)
	i++
	res[i] = fmt.Sprintf("%.4f", t.MinusDI)

	i++
	res[i] = fmt.Sprintf("%.4f", t.CCI14)

	i++
	res[i] = fmt.Sprintf("%.4f", t.BollUpper)
	i++
	res[i] = fmt.Sprintf("%.4f", t.BollMiddle)
	i++
	res[i] = fmt.Sprintf("%.4f", t.BollLower)

	i++
	res[i] = fmt.Sprintf("%.4f", t.MA5)
	i++
	res[i] = fmt.Sprintf("%.4f", t.MA10)
	i++
	res[i] = fmt.Sprintf("%.4f", t.MA20)
	i++
	res[i] = fmt.Sprintf("%.4f", t.MA60)
	i++
	res[i] = fmt.Sprintf("%.4f", t.EMA12)
	i++
	res[i] = fmt.Sprintf("%.4f", t.EMA26)

	i++
	res[i] = fmt.Sprintf("%.4f", t.ATR14)

	i++
	res[i] = fmt.Sprintf("%.4f", t.K)
	i++
	res[i] = fmt.Sprintf("%.4f", t.D)
	i++
	res[i] = fmt.Sprintf("%.4f", t.J)

	return res
}

type TechnicalIndicators struct {
	Close    *decimal.Decimal `toon:"close" json:"close"`
	Open     *decimal.Decimal `toon:"open" json:"open"`
	Low      *decimal.Decimal `toon:"low" json:"low"`
	High     *decimal.Decimal `toon:"high" json:"high"`
	Volume   *decimal.Decimal `toon:"volume" json:"volume"`     // 成交量
	Turnover *decimal.Decimal `toon:"turnover" json:"turnover"` // 成交额
	// 时间
	Timestamp int64     `toon:"timestamp" json:"timestamp"`
	Time      time.Time `toon:"time" json:"time"`
	TimeStr   string    `toon:"timeStr" json:"timeStr"`
	// MACD指标
	MACD       float64 `toon:"MACD" json:"MACD"`             // MACD线 (DIF)
	MACDSignal float64 `toon:"MACDSignal" json:"MACDSignal"` // 信号线 (DEA)
	MACDHist   float64 `toon:"MACDHist" json:"MACDHist"`     // 柱状图 (MACD柱)
	// RSI指标 (多周期)
	RSI6  float64 `toon:"RSI6" json:"RSI6"`
	RSI14 float64 `toon:"RSI14" json:"RSI14"`
	RSI24 float64 `toon:"RSI24" json:"RSI24"`
	// ADX指标 (趋势强度)
	ADX     float64 `toon:"ADX" json:"ADX"`
	PlusDI  float64 `toon:"plusDI" json:"plusDI"`   // +DI
	MinusDI float64 `toon:"minusDI" json:"minusDI"` // -DI
	// MFI指标 (资金流量)
	MFI14 float64 `toon:"MFI14" json:"MFI14"`
	// CCI指标
	CCI14 float64 `toon:"CCI14" json:"CCI14"`
	CCI20 float64 `toon:"CCI20" json:"CCI20"`
	// 动量指标
	MOM10 float64 `toon:"MOM10" json:"MOM10"`
	ROC10 float64 `toon:"ROC10" json:"ROC10"` // 变化率
	// 累积/派发线 (AD Line)
	AD float64 `toon:"AD" json:"AD"`
	// 布林带
	BollUpper  float64 `toon:"BollUpper" json:"BollUpper"`
	BollMiddle float64 `toon:"BollMiddle" json:"BollMiddle"`
	BollLower  float64 `toon:"BollLower" json:"BollLower"`
	// 均线
	MA5   float64 `toon:"MA5" json:"MA5"`
	MA10  float64 `toon:"MA10" json:"MA10"`
	MA20  float64 `toon:"MA20" json:"MA20"`
	MA60  float64 `toon:"MA60" json:"MA60"`
	EMA12 float64 `toon:"EMA12" json:"EMA12"`
	EMA26 float64 `toon:"EMA26" json:"EMA26"`
	// ATR (平均真实波幅)
	ATR14 float64 `toon:"ATR14" json:"ATR14"`
	// KDJ指标
	K float64 `toon:"K" json:"K"`
	D float64 `toon:"D" json:"D"`
	J float64 `toon:"J" json:"J"`
	// OBV (能量潮)
	OBV float64 `toon:"OBV" json:"OBV"`
}
