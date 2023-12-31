package job

import (
	"context"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/entity"
	"github.com/fadyat/i4u/pkg/syncs"
	"go.uber.org/zap"
	"time"
)

type producer struct {
	mailClient     api.Mail
	analyzerClient api.Analyzer
	summarizer     api.Summarizer
	sender         api.Sender

	labelsMapper *config.LabelsMapper
}

func NewProducer(
	mailClient api.Mail,
	analyzerClient api.Analyzer,
	summarizer api.Summarizer,
	sender api.Sender,
	labelsMapper *config.LabelsMapper,
) Producer {
	return &producer{
		mailClient:     mailClient,
		analyzerClient: analyzerClient,
		summarizer:     summarizer,
		sender:         sender,
		labelsMapper:   labelsMapper,
	}
}

func (p *producer) Produce(ctx context.Context) <-chan error {
	errsCh := make(chan error)
	labelerChan := make(chan entity.Message)
	analyzerChan := make(chan entity.Message)
	summarizerChan := make(chan entity.Message)
	senderChan := make(chan entity.SummaryMsg)

	fetcherJob := NewFetcherJob(
		p.mailClient,
		10*time.Second,
		errsCh,
		[]chan<- entity.Message{labelerChan, analyzerChan},
	)
	labelerJob := NewLabelerJob(p.mailClient, errsCh, labelerChan)
	analyzerJob := NewAnalyzerJob(
		p.analyzerClient,
		p.labelsMapper,
		errsCh,
		analyzerChan,
		[]chan<- entity.Message{summarizerChan, labelerChan},
	)
	summarizerJob := NewSummarizerJob(p.summarizer, errsCh, summarizerChan, senderChan)
	senderJob := NewSenderJob(p.sender, errsCh, senderChan)

	var (
		jobsWg syncs.WaitGroup
		jobs   = []Job{
			fetcherJob, labelerJob, analyzerJob, summarizerJob, senderJob,
		}
	)

	for _, j := range jobs {
		job := j

		jobsWg.Go(func() {
			zap.S().Infof("starting %T", job)
			job.Run(ctx)
		})
	}

	go func() {
		defer func() {
			zap.S().Info("stopping producer and all channels")
			close(errsCh)
			close(labelerChan)
			close(analyzerChan)
			close(summarizerChan)
			close(senderChan)
		}()

		<-ctx.Done()
		jobsWg.Wait()
	}()

	return errsCh
}
