package opentracing

import "github.com/uber/jaeger-client-go"

type TagSampler struct {
	tagValues map[string][]interface{}
}

type TagValue struct {
	Tag   string
	Value interface{}
}

var (
	undecidedDecision = jaeger.SamplingDecision{Sample: false, Retryable: true, Tags: nil}
	sampleDecision    = jaeger.SamplingDecision{
		Sample:    true,
		Retryable: false,
		Tags: []jaeger.Tag{
			jaeger.NewTag("sampler.type", "TagSampler"),
		},
	}
)

func NewTagSampler(matches []TagValue) *TagSampler {
	tv := make(map[string][]interface{})
	for _, match := range matches {
		tv[match.Tag] = append(tv[match.Tag], match.Value)
	}
	return &TagSampler{
		tagValues: tv,
	}
}

func (t *TagSampler) OnCreateSpan(span *jaeger.Span) jaeger.SamplingDecision {
	return undecidedDecision
}

func (t *TagSampler) OnSetOperationName(span *jaeger.Span, operationName string) jaeger.SamplingDecision {
	return undecidedDecision
}

func (t *TagSampler) OnSetTag(span *jaeger.Span, key string, value interface{}) jaeger.SamplingDecision {
	if values, found := t.tagValues[key]; found {
		for _, mvalue := range values {
			if mvalue == value {
				return sampleDecision
			}
		}
	}
	return undecidedDecision
}

func (s *TagMatchingSampler) OnFinishSpan(span *jaeger.Span) jaeger.SamplingDecision {
	return undecidedDecision
}
