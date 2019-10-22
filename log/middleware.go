package log

import (
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

// Logger interface defines the behaviour of a log instance. Every logger
// requires to implement these methods to be used by the Middleware
type Logger interface {
	Printf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Middleware implementations io.Writer to capture all the Terraform logs using
// the "log" package and send them to the defined logger
type Middleware struct {
	log        Logger
	prevWriter io.Writer
	mu         sync.Mutex
}

// NewMiddleware creates a new instance of Middleware with the Standard
// Logger (if nil) a the given logger
func NewMiddleware(l Logger) *Middleware {
	// If there is no logger, use the Terranova log wrapper
	if l == nil {
		l = NewLog(os.Stdout, "", DefLogLevel)
	}

	m := &Middleware{
		log: l,
	}

	// Redirect the Terraform log output to the middleware. Then Write() will send
	// the right output to the received/created logger
	prevWriter := log.Writer()
	log.SetOutput(m)
	m.prevWriter = prevWriter

	return m
}

// Close restore the output of the standard logger (used by Terraform) and stop
// using the Middleware
func (m *Middleware) Close() {
	log.SetOutput(m.prevWriter)
	m.log = nil
}

// SetLogger sets or changes the logger of the Middleware, which is the logger
// to send the Terraform output
func (m *Middleware) SetLogger(l Logger) {
	m.log = l
}

// Writer captures all the output from Terraform and use the logger to print it out
func (m *Middleware) Write(p []byte) (n int, err error) {
	// The regexp search for a timestamp, a label and the log message. Example:
	// 2019/10/20 20:43:00 [DEBUG] this is a debugging message
	re := regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+\[(\w+)\]\s+((?s:.+))`)
	allMatch := re.FindAllStringSubmatch(string(p), -1)

	if len(allMatch) > 0 {
		match := allMatch[0]
		logMessage := strings.TrimRight(match[2], "\n")
		if len(match) == 3 {
			switch match[1] {
			case "ERROR":
				m.log.Errorf(logMessage)
			case "WARN":
				m.log.Warnf(logMessage)
			case "INFO":
				m.log.Infof(logMessage)
			case "DEBUG":
				m.log.Debugf(logMessage)
			case "TRACE":
				m.log.Debugf("[LEVEL 2] %s", logMessage)
			default:
				// If there is a log entry with an unknown label (i.e. "[X]") in cyan,
				// improve the log output including the missing label in the switch/case
				m.log.Printf("\x1B[36m[%s]\x1B[0m (Unknown Log Label) %s", match[1], logMessage)
			}
		} else {
			// If you see something starting with ">", improve the log output
			// including the missing pattern in the regex
			m.log.Printf("> %s", p)
		}
	} else {
		// The regexp search for a timestamp and log message
		reDate := regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+(.+)`)
		allMatchDate := reDate.FindAllStringSubmatch(string(p), -1)
		matchDate := allMatchDate[0]
		if len(matchDate) == 2 {
			m.log.Printf(matchDate[1])
		} else {
			// If you see something starting with ">>", improve the log output
			// including the missing pattern in the regexp
			m.log.Printf(">> %s", p)
		}
	}

	return len(p), nil
}