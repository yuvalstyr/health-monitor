package charts

import (
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
)

// ChartData represents the data needed to generate a chart
type ChartData struct {
	Labels []string  `json:"labels"`
	Values []float64 `json:"values"`
	Target float64   `json:"target"`
	Unit   string    `json:"unit"`
	Title  string    `json:"title"`
}

// sanitizeChartID ensures the chartID contains only safe characters for HTML/JS
func sanitizeChartID(chartID string) string {
	// Allow only alphanumeric characters, hyphens, and underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	sanitized := reg.ReplaceAllString(chartID, "")
	
	// Ensure it starts with a letter (required for valid HTML ID)
	if len(sanitized) == 0 || !regexp.MustCompile(`^[a-zA-Z]`).MatchString(sanitized) {
		sanitized = "chart_" + sanitized
	}
	
	return sanitized
}

// EChartsOption represents the configuration for ECharts
type EChartsOption struct {
	BackgroundColor   string            `json:"backgroundColor"`
	Title             TitleOption       `json:"title"`
	Tooltip           TooltipOption     `json:"tooltip"`
	Legend            LegendOption      `json:"legend"`
	XAxis             XAxisOption       `json:"xAxis"`
	YAxis             YAxisOption       `json:"yAxis"`
	Series            []SeriesOption    `json:"series"`
	Grid              GridOption        `json:"grid"`
	TextStyle         TextStyleOption   `json:"textStyle"`
	Animation         bool              `json:"animation"`
	AnimationDuration int               `json:"animationDuration"`
	AnimationEasing   string            `json:"animationEasing"`
}

type TitleOption struct {
	Text      string          `json:"text"`
	TextStyle TextStyleOption `json:"textStyle"`
	Left      string          `json:"left,omitempty"`
	Top       string          `json:"top,omitempty"`
}

type TooltipOption struct {
	Trigger         string            `json:"trigger"`
	BackgroundColor string            `json:"backgroundColor"`
	BorderColor     string            `json:"borderColor"`
	BorderWidth     string            `json:"borderWidth,omitempty"`
	TextStyle       TextStyleOption   `json:"textStyle"`
	AxisPointer     AxisPointerOption `json:"axisPointer,omitempty"`
}

type LegendOption struct {
	Data      []string        `json:"data"`
	TextStyle TextStyleOption `json:"textStyle"`
	Top       string          `json:"top,omitempty"`
	Left      string          `json:"left,omitempty"`
}

type XAxisOption struct {
	Type      string            `json:"type"`
	Data      []string          `json:"data"`
	AxisLabel AxisLabelOption   `json:"axisLabel"`
	AxisLine  AxisLineOption    `json:"axisLine"`
	AxisTick  AxisTickOption    `json:"axisTick,omitempty"`
}

type YAxisOption struct {
	Type      string            `json:"type"`
	AxisLabel AxisLabelOption   `json:"axisLabel"`
	AxisLine  AxisLineOption    `json:"axisLine"`
	SplitLine SplitLineOption   `json:"splitLine"`
	AxisTick  AxisTickOption    `json:"axisTick,omitempty"`
}

type SeriesOption struct {
	Name       string             `json:"name"`
	Type       string             `json:"type"`
	Data       []float64          `json:"data"`
	Smooth     bool               `json:"smooth"`
	ItemStyle  ItemStyleOption    `json:"itemStyle"`
	LineStyle  LineStyleOption    `json:"lineStyle,omitempty"`
	Symbol     string             `json:"symbol,omitempty"`
	SymbolSize int                `json:"symbolSize,omitempty"`
	ShowSymbol bool               `json:"showSymbol,omitempty"`
	Emphasis   EmphasisOption     `json:"emphasis,omitempty"`
}

type GridOption struct {
	Left            string `json:"left"`
	Right           string `json:"right"`
	Top             string `json:"top"`
	Bottom          string `json:"bottom"`
	ContainLabel    bool   `json:"containLabel,omitempty"`
	BorderColor     string `json:"borderColor,omitempty"`
	BorderWidth     string `json:"borderWidth,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
}

type TextStyleOption struct {
	Color      string `json:"color"`
	FontSize   string `json:"fontSize,omitempty"`
	FontWeight string `json:"fontWeight,omitempty"`
}

type AxisLabelOption struct {
	Color    string `json:"color"`
	FontSize string `json:"fontSize,omitempty"`
}

type AxisLineOption struct {
	LineStyle LineStyleOption `json:"lineStyle"`
}

type SplitLineOption struct {
	LineStyle LineStyleOption `json:"lineStyle"`
}

type ItemStyleOption struct {
	Color       string `json:"color"`
	BorderColor string `json:"borderColor,omitempty"`
	BorderWidth string `json:"borderWidth,omitempty"`
}

type LineStyleOption struct {
	Color   string  `json:"color"`
	Type    string  `json:"type,omitempty"`
	Width   string  `json:"width,omitempty"`
	Opacity float64 `json:"opacity,omitempty"`
}

// New types for enhanced styling
type AreaStyleOption struct {
	Color interface{} `json:"color,omitempty"`
}

type EmphasisOption struct {
	ItemStyle ItemStyleOption `json:"itemStyle,omitempty"`
	Scale     bool            `json:"scale,omitempty"`
}

type AxisPointerOption struct {
	Type       string         `json:"type,omitempty"`
	CrossStyle CrossStyleOption `json:"crossStyle,omitempty"`
}

type CrossStyleOption struct {
	Color   string  `json:"color,omitempty"`
	Width   int     `json:"width,omitempty"`
	Opacity float64 `json:"opacity,omitempty"`
}

type AxisTickOption struct {
	Show      bool            `json:"show,omitempty"`
	LineStyle LineStyleOption `json:"lineStyle,omitempty"`
}

// RenderChartHTML renders the chart as HTML string  
func RenderChartHTML(data ChartData, chartID string) template.HTML {
	if len(data.Labels) == 0 || len(data.Values) == 0 {
		return template.HTML(GenerateEmptyChartHTML(data.Title))
	}

	// Sanitize chartID to prevent XSS
	safeChartID := sanitizeChartID(chartID)

	chartOption := generateEChartsOption(data)
	optionJSON, err := json.Marshal(chartOption)
	if err != nil {
		return template.HTML(GenerateEmptyChartHTML("Error: Failed to generate chart configuration"))
	}

	chartHTML := fmt.Sprintf(`
		<div id="%s" class="w-full h-96 bg-transparent chart-loading"></div>
		<script type="text/javascript">
			(function() {
				var chartDom = document.getElementById('%s');
				if (!chartDom || typeof echarts === 'undefined') {
					console.error('Chart container or ECharts library not found');
					return;
				}
				
				// Remove loading animation
				chartDom.classList.remove('chart-loading');
				
				// Initialize chart with no theme (blank) and transparent background
				var myChart = echarts.init(chartDom, null, {
					renderer: 'canvas',
					useDirtyRect: false
				});
				
				// Ensure transparent background on all elements
				myChart.getDom().style.backgroundColor = 'transparent';
				chartDom.style.backgroundColor = 'transparent';
				
				var option = %s;
				
				// Set chart options with animation
				myChart.setOption(option, true);
				
				// Force transparent background on all canvas elements
				setTimeout(function() {
					var canvases = chartDom.querySelectorAll('canvas');
					canvases.forEach(function(canvas) {
						canvas.style.backgroundColor = 'transparent';
					});
					// Also force the container
					chartDom.style.background = 'none';
					chartDom.style.backgroundColor = 'transparent';
				}, 100);
				
				// Enhanced resize handling with debouncing
				var resizeTimeout;
				function handleResize() {
					clearTimeout(resizeTimeout);
					resizeTimeout = setTimeout(function() {
						if (myChart && !myChart.isDisposed()) {
							myChart.resize();
						}
					}, 100);
				}
				
				window.addEventListener('resize', handleResize);
				
				// Add hover effects
				myChart.on('mouseover', function() {
					chartDom.parentElement.style.transform = 'scale(1.02)';
					chartDom.parentElement.style.transition = 'transform 0.2s ease-in-out';
				});
				
				myChart.on('mouseout', function() {
					chartDom.parentElement.style.transform = 'scale(1)';
				});
				
				// Store chart instance for cleanup
				window.chartInstances = window.chartInstances || {};
				window.chartInstances['%s'] = myChart;
			})();
		</script>
	`, safeChartID, safeChartID, string(optionJSON), safeChartID)

	return template.HTML(chartHTML)
}

// generateEChartsOption creates ECharts configuration from chart data
func generateEChartsOption(data ChartData) EChartsOption {
	// Create gradient colors for the main line
	series := []SeriesOption{
		{
			Name:   data.Title,
			Type:   "line",
			Data:   data.Values,
			Smooth: true,
			ItemStyle: ItemStyleOption{Color: "#570DF8"},
			LineStyle: LineStyleOption{
				Color: "#570DF8",
				Width: "3",
			},
			Symbol:     "circle",
			SymbolSize: 6,
			ShowSymbol: true,
			Emphasis: EmphasisOption{
				ItemStyle: ItemStyleOption{
					Color:      "#570DF8",
					BorderColor: "#ffffff",
					BorderWidth: "2",
				},
				Scale: true,
			},
		},
	}

	legendData := []string{data.Title}

	// Add target line if specified with enhanced styling
	if data.Target > 0 {
		targetData := make([]float64, len(data.Values))
		for i := range targetData {
			targetData[i] = data.Target
		}

		targetSeries := SeriesOption{
			Name:   fmt.Sprintf("Target (%.1f %s)", data.Target, data.Unit),
			Type:   "line",
			Data:   targetData,
			Smooth: false,
			ItemStyle: ItemStyleOption{Color: "#F87272"},
			LineStyle: LineStyleOption{
				Color: "#F87272",
				Type:  "dashed",
				Width: "2",
			},
			Symbol:     "none",
			SymbolSize: 0,
		}

		series = append(series, targetSeries)
		legendData = append(legendData, fmt.Sprintf("Target (%.1f %s)", data.Target, data.Unit))
	}

	return EChartsOption{
		BackgroundColor: "transparent",
		TextStyle:       TextStyleOption{Color: "#a6adba"},
		Title: TitleOption{
			Text: data.Title,
			TextStyle: TextStyleOption{
				Color:      "#ffffff",
				FontSize:   "18",
				FontWeight: "bold",
			},
			Left: "center",
			Top:  "2%",
		},
		Tooltip: TooltipOption{
			Trigger:         "axis",
			BackgroundColor: "transparent",
			BorderColor:     "transparent",
			TextStyle:       TextStyleOption{Color: "#ffffff"},
			AxisPointer: AxisPointerOption{
				Type: "cross",
				CrossStyle: CrossStyleOption{
					Color: "#570DF8",
					Width: 1,
					Opacity: 0.6,
				},
			},
		},
		Legend: LegendOption{
			Data:      legendData,
			TextStyle: TextStyleOption{Color: "#a6adba"},
			Top:       "8%",
			Left:      "center",
		},
		XAxis: XAxisOption{
			Type: "category",
			Data: data.Labels,
			AxisLabel: AxisLabelOption{
				Color:    "#a6adba",
				FontSize: "12",
			},
			AxisLine: AxisLineOption{
				LineStyle: LineStyleOption{
					Color: "#374151",
					Width: "1",
				},
			},
			AxisTick: AxisTickOption{
				Show: true,
				LineStyle: LineStyleOption{
					Color: "#374151",
				},
			},
		},
		YAxis: YAxisOption{
			Type: "value",
			AxisLabel: AxisLabelOption{
				Color:    "#a6adba",
				FontSize: "12",
			},
			AxisLine: AxisLineOption{
				LineStyle: LineStyleOption{
					Color: "#374151",
					Width: "1",
				},
			},
			SplitLine: SplitLineOption{
				LineStyle: LineStyleOption{
					Color:   "#374151",
					Type:    "dashed",
					Opacity: 0.3,
				},
			},
			AxisTick: AxisTickOption{
				Show: true,
				LineStyle: LineStyleOption{
					Color: "#374151",
				},
			},
		},
		Series:    series,
		Animation: true,
		AnimationDuration: 1000,
		AnimationEasing:   "cubicOut",
		Grid: GridOption{
			Left:           "8%",
			Right:          "8%",
			Top:            "20%",
			Bottom:         "15%",
			ContainLabel:   true,
		},
	}
}

// GenerateEmptyChartHTML creates HTML for empty chart state
func GenerateEmptyChartHTML(title string) string {
	return fmt.Sprintf(`
		<div class="flex items-center justify-center h-96 bg-base-200 rounded-lg border-2 border-dashed border-base-300">
			<div class="text-center">
				<svg class="mx-auto h-12 w-12 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 00-2-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
				</svg>
				<h3 class="mt-2 text-sm font-medium text-base-content/70">No data available for %s</h3>
			</div>
		</div>
	`, title)
}