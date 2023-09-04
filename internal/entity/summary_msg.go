package entity

type SummaryMessage interface {
	Summary() string
}

type SummaryMsg struct {
	Message
	summary string
}

type AlertMsg struct {
	err error
}

func NewAlertMsg(err error) *AlertMsg {
	return &AlertMsg{err: err}
}

func (a *AlertMsg) Summary() string {
	return a.err.Error()
}

func NewSummaryMsg(msg Message, summary string) *SummaryMsg {
	return &SummaryMsg{
		Message: msg,
		summary: summary,
	}
}

func (s *SummaryMsg) Summary() string {
	return s.summary + "\n\n" + s.Message.Link()
}
