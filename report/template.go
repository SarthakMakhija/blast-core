package report

import (
	"fmt"
	"io"
	"text/template"
	"time"
)

var templateText = `
Summary:
  LoadMetrics:
    TotalRequests: {{ formatNumberUint .Load.TotalRequests }}
    SuccessCount: {{ formatNumberUint .Load.SuccessCount }}
    ErrorCount: {{ formatNumberUint .Load.ErrorCount }}
    TotalPayloadSize: {{ formatNumberInt64 .Load.TotalPayloadLengthBytes }} bytes
    AveragePayloadSize: {{ formatNumberFloat .Load.AveragePayloadLengthBytes }} bytes
    EarliestLoadSendTime: {{ formatTime .Load.EarliestLoadSendTime}}
    LatestLoadSendTime: {{ formatTime .Load.LatestLoadSendTime}}
    TimeToCompleteLoad: {{ formatDuration .Load.TotalTime }}

{{ if gt (len .Load.ErrorCountByType) 0 }}  Error distribution:{{ range $err, $num := .Load.ErrorCountByType }}
  [{{ $num }}]   {{ $err }}{{ end }}{{ end }}

{{ if eq (.Response.IsAvailableForReporting) true }}  
  ResponseMetrics:
    SuccessCount: {{ formatNumberUint .Response.SuccessCount }}
    ErrorCount: {{ formatNumberUint .Response.ErrorCount }}
    TotalResponsePayloadSize: {{ formatNumberInt64 .Response.TotalResponsePayloadLengthBytes }} bytes
    AverageResponsePayloadSize: {{ formatNumberFloat .Response.AverageResponsePayloadLengthBytes }} bytes
    EarliestResponseReceivedTime: {{ formatTime .Response.EarliestResponseReceivedTime }}
    LatestResponseReceivedTime: {{ formatTime .Response.LatestResponseReceivedTime }}
    TimeToGetResponses: {{ formatDuration .Response.TotalTime }}
  
{{ if gt (len .Response.ErrorCountByType) 0 }}  Error distribution:{{ range $err, $num := .Response.ErrorCountByType }} 
  [{{ $num }}]   {{ $err }}{{ end }}{{ end }}
{{end}}
`

var functions = template.FuncMap{
	"formatNumberFloat": formatNumberFloat,
	"formatNumberUint":  formatNumberUint,
	"formatNumberInt64": formatNumberInt64,
	"formatTime":        formatTime,
	"formatDuration":    formatDuration,
}

const timeFormat = "January 02, 2006 15:04:05 MST"

func formatNumberFloat(value float64) string {
	return fmt.Sprintf("%4.4f", value)
}

func formatNumberUint(value uint) string {
	return fmt.Sprintf("%d", value)
}

func formatNumberInt64(value int64) string {
	return fmt.Sprintf("%d", value)
}

func formatTime(time time.Time) string {
	return time.Format(timeFormat)
}

func formatDuration(duration time.Duration) string {
	return duration.String()
}

func print(writer io.Writer, report *Report) error {
	return newTemplate().Execute(writer, report)
}

func newTemplate() *template.Template {
	return template.Must(template.New("blast").Funcs(functions).Parse(templateText))
}
