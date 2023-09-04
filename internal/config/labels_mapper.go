package config

// LabelsMapper is a struct that holds the labels mapping
// between the labels that will be created in your Gmail account,
// and the labels that will be used for marking processed messages
// to avoid processing them again.
type LabelsMapper struct {
	I4U       string `yaml:"i4u"`
	NotIntern string `yaml:"intern:false"`
	IsIntern  string `yaml:"intern:true"`
}

func (l *LabelsMapper) GetInternLabel(isIntern bool) string {
	if isIntern {
		return l.IsIntern
	}

	return l.NotIntern
}
