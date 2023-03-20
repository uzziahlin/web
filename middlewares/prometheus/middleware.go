package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/uzziahlin/web"
	"strconv"
	"time"
)

type MiddlewareBuilder struct {
	Name        string
	Subsystem   string
	ConstLabels map[string]string
	Help        string
}

func (m *MiddlewareBuilder) Build() web.Middleware {
	summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:        m.Name,
		Subsystem:   m.Subsystem,
		ConstLabels: m.ConstLabels,
		Help:        m.Help,
	}, []string{"pattern", "method", "status"})
	prometheus.MustRegister(summaryVec)
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			startTime := time.Now()
			next(ctx)
			endTime := time.Now()
			go report(endTime.Sub(startTime), ctx, summaryVec)
		}
	}
}

func report(dur time.Duration, ctx *web.Context, vec prometheus.ObserverVec) {
	status := ctx.RespStatus
	route := "unknown"
	if ctx.MatchedRoute != "" {
		route = ctx.MatchedRoute
	}
	ms := dur / time.Millisecond
	vec.WithLabelValues(route, ctx.Req.Method, strconv.Itoa(status)).Observe(float64(ms))
}
