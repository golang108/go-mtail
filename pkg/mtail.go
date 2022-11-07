package mtail

import (
	"context"
	"github.com/google/mtail/internal/metrics"
	"github.com/google/mtail/internal/mtail"
	"github.com/google/mtail/internal/waker"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

func NewMtailRegistry(ctx context.Context) (_ *prometheus.Registry, err error) {
	progs := ""
	logs := []string{"1", "2"}
	ignoreFileRegPattern := ""
	buildInfo := mtail.BuildInfo{
		Version:  "version",
		Branch:   "branch",
		Revision: "Revision",
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	metricPushInterval := 1 * time.Minute
	maxRegexpLen := 1024
	maxRecursionDepth := 100

	opts := []mtail.Option{
		mtail.ProgramPath(progs),
		mtail.LogPathPatterns(logs...),
		mtail.IgnoreRegexPattern(ignoreFileRegPattern),
		mtail.SetBuildInfo(buildInfo),
		mtail.OverrideLocation(loc),
		mtail.MetricPushInterval(metricPushInterval),
		mtail.MaxRegexpLength(maxRegexpLen),
		mtail.MaxRecursionDepth(maxRecursionDepth),
		mtail.LogRuntimeErrors,
	}

	staleLogGcWaker := waker.NewTimed(ctx, time.Hour)
	opts = append(opts, mtail.StaleLogGcWaker(staleLogGcWaker))

	pollInterval := 250 * time.Millisecond
	pollLogInterval :=  250 * time.Millisecond
	if pollInterval > 0 {
		logStreamPollWaker := waker.NewTimed(ctx, pollInterval)
		logPatternPollWaker := waker.NewTimed(ctx, pollLogInterval)
		opts = append(opts, mtail.LogPatternPollWaker(logPatternPollWaker), mtail.LogstreamPollWaker(logStreamPollWaker))
	}
	sysLogUseCurrentYear := true
	if sysLogUseCurrentYear {
		opts = append(opts, mtail.SyslogUseCurrentYear)
	}

	emitProgLabel := true
	if emitProgLabel {
		opts = append(opts, mtail.OmitProgLabel)
	}

	emitMetricTimestamp := true
	if emitMetricTimestamp {
		opts = append(opts, mtail.EmitMetricTimestamp)
	}

	store := metrics.NewStore()
	store.StartGcLoop(ctx, time.Hour)

	m, err := mtail.New(ctx, store, opts...)
	if err != nil {
		return
	}
	reg := m.GetRegistry()

	return reg, nil
}
