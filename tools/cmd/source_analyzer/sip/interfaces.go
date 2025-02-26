// github.com/delving/hub3/tools/cmd/source_analyzer/sip/interfaces.go
package sip

// AnalysisListener handles analysis results and errors
type AnalysisListener interface {
	// Success is called when analysis completes successfully
	Success(stats *Stats)
	// Failure is called when analysis encounters an error
	Failure(message string, err error)
}

// ProgressListener monitors analysis progress
type ProgressListener interface {
	// SetProgress reports the number of elements processed
	SetProgress(count int)
	// SetProgressMessage updates the current processing status
	SetProgressMessage(message string)
}

// DefaultAnalysisListener provides a basic implementation of AnalysisListener
type DefaultAnalysisListener struct{}

func (l *DefaultAnalysisListener) Success(stats *Stats)              {}
func (l *DefaultAnalysisListener) Failure(message string, err error) {}

// DefaultProgressListener provides a basic implementation of ProgressListener
type DefaultProgressListener struct{}

func (l *DefaultProgressListener) SetProgress(count int)             {}
func (l *DefaultProgressListener) SetProgressMessage(message string) {}
