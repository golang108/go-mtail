package mtail

import (
	"context"
	"github.com/google/mtail/internal/metrics"
	"github.com/google/mtail/internal/mtail"
	"github.com/google/mtail/internal/waker"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type Option struct {
	NamePrefix string
	Version    string
	Progs      string
	Logs       []string

	IgnoreFileRegPattern string
	OverrideTimeZone     string
	OmitProgLabel        bool
	EmitMetricTimestamp  bool
	PollInterval         time.Duration
	PollLogInterval      time.Duration
	MetricPushInterval   time.Duration
	MaxRegexpLen         int
	MaxRecursionDepth    int
	SyslogUseCurrentYear bool
	LogRuntimeErrors     bool
}

func GetRegistry(ctx context.Context, ins Option) (_ *prometheus.Registry, err error) {
	buildInfo := mtail.BuildInfo{
		Version: ins.Version,
	}
	loc, err := time.LoadLocation(ins.OverrideTimeZone)
	if err != nil {
		return nil, err
	}

	opts := []mtail.Option{
		mtail.ProgramPath(ins.Progs),
		mtail.LogPathPatterns(ins.Logs...),
		mtail.IgnoreRegexPattern(ins.IgnoreFileRegPattern),
		mtail.SetBuildInfo(buildInfo),
		mtail.OverrideLocation(loc),
		mtail.MetricPushInterval(ins.MetricPushInterval), // keep it here ?
		mtail.MaxRegexpLength(ins.MaxRegexpLen),
		mtail.MaxRecursionDepth(ins.MaxRecursionDepth),
		mtail.LogRuntimeErrors,
	}

	staleLogGcWaker := waker.NewTimed(ctx, time.Hour)
	opts = append(opts, mtail.StaleLogGcWaker(staleLogGcWaker))

	if ins.PollInterval > 0 {
		logStreamPollWaker := waker.NewTimed(ctx, ins.PollInterval)
		logPatternPollWaker := waker.NewTimed(ctx, ins.PollLogInterval)
		opts = append(opts, mtail.LogPatternPollWaker(logPatternPollWaker), mtail.LogstreamPollWaker(logStreamPollWaker))
	}
	if ins.SyslogUseCurrentYear {
		opts = append(opts, mtail.SyslogUseCurrentYear)
	}
	if ins.OmitProgLabel {
		opts = append(opts, mtail.OmitProgLabel)
	}
	if ins.EmitMetricTimestamp {
		opts = append(opts, mtail.EmitMetricTimestamp)
	}

	store := metrics.NewStore()
	store.StartGcLoop(ctx, time.Hour)

	m, err := mtail.NewMtail(ctx, store, opts...)
	if err != nil {
		return
	}
	reg := m.GetRegistry()

	return reg, nil

}
