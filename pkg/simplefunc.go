package pkg

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"text/template"
	"time"
)

// simpleFuncMap provides basic template functions without heavy dependencies
func SimpleFuncMap() template.FuncMap {
	return template.FuncMap{
		"toJSON": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			return string(b), err
		},
		"toMap": func(v interface{}) (map[string]interface{}, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			var m map[string]interface{}
			err = json.Unmarshal(b, &m)
			return m, err
		},
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
		},
		"now": func() time.Time {
			return time.Now()
		},
		"unixEpoch": func() int64 {
			return time.Now().Unix()
		},
		"randAlphaNum": func(length int) string {
			const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			b := make([]byte, length)
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := range b {
				b[i] = charset[r.Intn(len(charset))]
			}
			return string(b)
		},
		"upper": func(s string) string {
			return s // Simple version, for now just return as-is
		},
		"date": func(format string, t time.Time) string {
			// Go time format: convert common formats
			switch format {
			case "2006-01-02T15:04:05Z07:00":
				return t.Format("2006-01-02T15:04:05Z07:00")
			case "2006-01-02":
				return t.Format("2006-01-02")
			default:
				return t.Format(format)
			}
		},
		"debug": func(v interface{}) string {
			return fmt.Sprintf("%#v", v)
		},
	}
}
