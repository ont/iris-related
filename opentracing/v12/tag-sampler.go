package opentracing

import (
	"regexp"

	"github.com/uber/jaeger-client-go"
)

type TagSampler struct {
	tagMatches map[string][]TagMatch
}

type TagMatch struct {
	Tag      string
	Matcher  Matcher
	Decision SamplingDecision
}

type SamplingDecision int

const (
	DecisionNextSampler SamplingDecision = iota
	DecisionTake
	DecisionDrop
)

type Matcher interface {
	Check(value interface{}) bool
}

type StringRegexpMatcher struct {
	r *regexp.Regexp
}

func NewStringRegexpMatcher(re string) *StringRegexpMatcher {
	return &StringRegexpMatcher{
		r: regexp.MustCompile(re),
	}
}

func (m *StringRegexpMatcher) Check(value interface{}) bool {
	if sval, ok := value.(string); ok {
		return m.r.MatchString(sval)
	}
	return false
}

type MatchAllMatcher struct{}

func (m *MatchAllMatcher) Check(value interface{}) bool {
	return true
}

var (
	undecidedDecision = jaeger.SamplingDecision{Sample: false, Retryable: true, Tags: nil}
	notSampleDecision = jaeger.SamplingDecision{
		Sample:    false,
		Retryable: false,
		Tags: []jaeger.Tag{
			jaeger.NewTag("sampler.type", "TagSampler"),
		},
	}
	sampleDecision = jaeger.SamplingDecision{
		Sample:    true,
		Retryable: false,
		Tags: []jaeger.Tag{
			jaeger.NewTag("sampler.type", "TagSampler"),
		},
	}
)

func NewTagSampler(matches []TagMatch) *TagSampler {
	tm := make(map[string][]TagMatch)
	for _, match := range matches {
		tm[match.Tag] = append(tm[match.Tag], match)
	}
	return &TagSampler{
		tagMatches: tm,
	}
}

func (t *TagSampler) OnCreateSpan(span *jaeger.Span) jaeger.SamplingDecision {
	return undecidedDecision
}

func (t *TagSampler) OnSetOperationName(span *jaeger.Span, operationName string) jaeger.SamplingDecision {
	return undecidedDecision
}

func (t *TagSampler) OnSetTag(span *jaeger.Span, key string, value interface{}) jaeger.SamplingDecision {
	if matches, found := t.tagMatches[key]; found {
		for _, match := range matches {
			if match.Matcher.Check(value) {
				return t.jaegerDecision(match.Decision)
			}
		}
	}
	return undecidedDecision
}

func (s *TagSampler) OnFinishSpan(span *jaeger.Span) jaeger.SamplingDecision {
	return undecidedDecision
}

func (s *TagSampler) Close() {
	return
}

func (t *TagSampler) jaegerDecision(decision SamplingDecision) jaeger.SamplingDecision {
	switch decision {
	case DecisionDrop:
		return notSampleDecision
	case DecisionTake:
		return sampleDecision
	case DecisionNextSampler:
		return undecidedDecision
	default:
		return undecidedDecision
	}
}
