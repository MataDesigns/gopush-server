package gopushserver

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)

// LogReq is http request log
type LogReq struct {
	URI         string `json:"uri"`
	Method      string `json:"method"`
	IP          string `json:"ip"`
	ContentType string `json:"content_type"`
	Agent       string `json:"agent"`
}

// LogPushEntry is push response log
type LogPushEntry struct {
	Type     string `json:"type"`
	Platform string `json:"platform"`
	Token    string `json:"token"`
	Message  string `json:"message"`
	Error    string `json:"error"`
}

var isTerm bool

func init() {
	isTerm = isatty.IsTerminal(os.Stdout.Fd())
}

// InitLog use for initial log module
func InitLog() {
	// init logger
	LogAccess = logrus.New()
	LogError = logrus.New()

	LogAccess.Formatter = &logrus.TextFormatter{
		TimestampFormat: "2006/01/02 - 15:04:05",
		FullTimestamp:   true,
	}

	LogError.Formatter = &logrus.TextFormatter{
		TimestampFormat: "2006/01/02 - 15:04:05",
		FullTimestamp:   true,
	}
}

// SetLogOut provide log stdout and stderr output
func SetLogOut(log *logrus.Logger, outString string) error {
	switch outString {
	case "stdout":
		log.Out = os.Stdout
	case "stderr":
		log.Out = os.Stderr
	default:
		f, err := os.OpenFile(outString, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

		if err != nil {
			return err
		}

		log.Out = f
	}

	return nil
}

// SetLogLevel is define log level what you want
// log level: panic, fatal, error, warn, info and debug
func SetLogLevel(log *logrus.Logger, levelString string) error {
	level, err := logrus.ParseLevel(levelString)

	if err != nil {
		return err
	}

	log.Level = level

	return nil
}

// LogRequest record http request
func LogRequest(uri string, method string, ip string, contentType string, agent string) {
	var output string
	log := &LogReq{
		URI:         uri,
		Method:      method,
		IP:          ip,
		ContentType: contentType,
		Agent:       agent,
	}

	logJSON, _ := json.Marshal(log)

	output = string(logJSON)

	LogAccess.Info(output)
}

func colorForPlatForm(platform int) string {
	switch platform {
	case PlatFormIos:
		return blue
	case PlatFormAndroid:
		return yellow
	default:
		return reset
	}
}

func typeForPlatForm(platform int) string {
	switch platform {
	case PlatFormIos:
		return "ios"
	case PlatFormAndroid:
		return "android"
	default:
		return ""
	}
}

func hideToken(token string, markLen int) string {
	if len(token) == 0 {
		return ""
	}

	if len(token) < markLen*2 {
		return strings.Repeat("*", len(token))
	}

	start := token[len(token)-markLen:]
	end := token[0:markLen]

	result := strings.Replace(token, start, strings.Repeat("*", markLen), -1)
	result = strings.Replace(result, end, strings.Repeat("*", markLen), -1)

	return result
}

func getLogPushEntry(status, token string, req PushNotification, errPush error) LogPushEntry {
	var errMsg string

	plat := typeForPlatForm(req.Platform)

	if errPush != nil {
		errMsg = errPush.Error()
	}

	return LogPushEntry{
		Type:     status,
		Platform: plat,
		Token:    token,
		Message:  req.Message,
		Error:    errMsg,
	}
}

// LogPush record user push request and server response.
func LogPush(status, token string, req PushNotification, errPush error) {
	var output string

	log := getLogPushEntry(status, token, req, errPush)
	logJSON, _ := json.Marshal(log)

	output = string(logJSON)

	switch status {
	case SucceededPush:
		LogAccess.Info(output)
	case FailedPush:
		LogError.Error(output)
	}
}

// LogMiddleware provide gin router handler.
func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		LogRequest(c.Request.URL.Path, c.Request.Method, c.ClientIP(), c.ContentType(), c.GetHeader("User-Agent"))
		c.Next()
	}
}
