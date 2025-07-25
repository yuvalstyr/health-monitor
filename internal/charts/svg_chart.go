package charts

import (
	"encoding/json"
	"fmt"
	"html/template"
)

// ChartData represents the data needed to generate a chart
type ChartData struct {
	Labels []string  `json:"labels"`
	Values []float64 `json:"values"`
	Target float64   `json:"target"`
	Unit   string    `json:"unit"`
	Title  string    `json:"title"`
}

// EChartsOption represents the configuration for ECharts
type EChartsOption struct {
	Title       TitleOption       `json:"title"`
	Tooltip     TooltipOption     `json:"tooltip"`
	Legend      LegendOption      `json:"legend"`
	XAxis       XAxisOption       `json:"xAxis"`
	YAxis       YAxisOption       `json:"yAxis"`
	Series      []SeriesOption    `json:"series"`
	Grid        GridOption        `json:"grid"`
	BackgroundColor string        `json:"backgroundColor"`
	TextStyle   TextStyleOption   `json:"textStyle"`
}

type TitleOption struct {
	Text      string          `json:"text"`
	TextStyle TextStyleOption `json:"textStyle"`
}

type TooltipOption struct {
	Trigger         string `json:"trigger"`
	BackgroundColor string `json:"backgroundColor"`
	BorderColor     string `json:"borderColor"`
	TextStyle       TextStyleOption `json:"textStyle"`
}

type LegendOption struct {
	Data      []string        `json:"data"`
	TextStyle TextStyleOption `json:"textStyle"`
}

type XAxisOption struct {
	Type      string            `json:"type"`
	Data      []string          `json:"data"`
	AxisLabel AxisLabelOption   `json:"axisLabel"`
	AxisLine  AxisLineOption    `json:"axisLine"`
}

type YAxisOption struct {
	Type      string            `json:"type"`
	AxisLabel AxisLabelOption   `json:"axisLabel"`
	AxisLine  AxisLineOption    `json:"axisLine"`
	SplitLine SplitLineOption   `json:"splitLine"`
}

type SeriesOption struct {
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	Data      []float64          `json:"data"`
	Smooth    bool               `json:"smooth"`
	ItemStyle ItemStyleOption    `json:"itemStyle"`
	LineStyle LineStyleOption    `json:"lineStyle,omitempty"`
}

type GridOption struct {
	Left   string `json:"left"`
	Right  string `json:"right"`
	Top    string `json:"top"`
	Bottom string `json:"bottom"`
}

type TextStyleOption struct {
	Color string `json:"color"`
}

type AxisLabelOption struct {
	Color string `json:"color"`
}

type AxisLineOption struct {
	LineStyle LineStyleOption `json:"lineStyle"`
}

type SplitLineOption struct {
	LineStyle LineStyleOption `json:"lineStyle"`
}

type ItemStyleOption struct {
	Color string `json:"color"`
}

type LineStyleOption struct {
	Color string `json:"color"`
	Type  string `json:"type,omitempty"`
}

// RenderChartHTML renders the chart as HTML string  
func RenderChartHTML(data ChartData, chartID string) template.HTML {
	if len(data.Labels) == 0 || len(data.Values) == 0 {
		return template.HTML(GenerateEmptyChartHTML(data.Title))
	}

	chartOption := generateEChartsOption(data)
	optionJSON, _ := json.Marshal(chartOption)

	chartHTML := fmt.Sprintf(`
		<div id="%s" class="w-full h-96 bg-transparent"></div>
		<script type="text/javascript">
			(function() {
				var chartDom = document.getElementById('%s');
				if (chartDom && typeof echarts !== 'undefined') {
					var myChart = echarts.init(chartDom, 'dark');
					var option = %s;
					myChart.setOption(option);
					
					// Resize chart on window resize
					window.addEventListener('resize', function() {
						myChart.resize();
					});
				}
			})();
		</script>
	`, chartID, chartID, string(optionJSON))

	return template.HTML(chartHTML)
}

// generateEChartsOption creates ECharts configuration from chart data
func generateEChartsOption(data ChartData) EChartsOption {
	series := []SeriesOption{
		{
			Name:      data.Title,
			Type:      "line",
			Data:      data.Values,
			Smooth:    true,
			ItemStyle: ItemStyleOption{Color: "#570DF8"},
		},
	}

	legendData := []string{data.Title}

	// Add target line if specified
	if data.Target > 0 {
		targetData := make([]float64, len(data.Values))
		for i := range targetData {
			targetData[i] = data.Target
		}

		targetSeries := SeriesOption{
			Name:      fmt.Sprintf("Target (%.1f %s)", data.Target, data.Unit),
			Type:      "line",
			Data:      targetData,
			Smooth:    false,
			ItemStyle: ItemStyleOption{Color: "#F87272"},
			LineStyle: LineStyleOption{Color: "#F87272", Type: "dashed"},
		}

		series = append(series, targetSeries)
		legendData = append(legendData, fmt.Sprintf("Target (%.1f %s)", data.Target, data.Unit))
	}

	return EChartsOption{
		BackgroundColor: "transparent",
		TextStyle:       TextStyleOption{Color: "#a6adba"},
		Title: TitleOption{
			Text:      data.Title,
			TextStyle: TextStyleOption{Color: "#ffffff"},
		},
		Tooltip: TooltipOption{
			Trigger:         "axis",
			BackgroundColor: "#1f2937",
			BorderColor:     "#374151",
			TextStyle:       TextStyleOption{Color: "#ffffff"},
		},
		Legend: LegendOption{
			Data:      legendData,
			TextStyle: TextStyleOption{Color: "#a6adba"},
		},
		XAxis: XAxisOption{
			Type:      "category",
			Data:      data.Labels,
			AxisLabel: AxisLabelOption{Color: "#a6adba"},
			AxisLine:  AxisLineOption{LineStyle: LineStyleOption{Color: "#374151"}},
		},
		YAxis: YAxisOption{
			Type:      "value",
			AxisLabel: AxisLabelOption{Color: "#a6adba"},
			AxisLine:  AxisLineOption{LineStyle: LineStyleOption{Color: "#374151"}},
			SplitLine: SplitLineOption{LineStyle: LineStyleOption{Color: "#374151"}},
		},
		Series: series,
		Grid: GridOption{
			Left:   "10%",
			Right:  "10%",
			Top:    "15%",
			Bottom: "15%",
		},
	}
}

// GenerateEmptyChartHTML creates HTML for empty chart state
func GenerateEmptyChartHTML(title string) string {
	return fmt.Sprintf(`
		<div class="flex items-center justify-center h-96 bg-gray-50 rounded-lg border-2 border-dashed border-gray-300">
			<div class="text-center">
				<svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 00-2-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
				</svg>
				<h3 class="mt-2 text-sm font-medium text-gray-900">No data available for %s</h3>
			</div>
		</div>
	`, title)
}