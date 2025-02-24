package util

import (
	"fmt"
	"github.com/fatih/color"
	"strings"
	"time"
)

func PrintColor(text string, colorType string) {
	var c *color.Color

	switch colorType {
	case "red":
		c = color.New(color.FgRed)
	case "green":
		c = color.New(color.FgGreen)
	case "yellow":
		c = color.New(color.FgYellow)
	case "blue":
		c = color.New(color.FgBlue)
	case "cyan":
		c = color.New(color.FgCyan)
	case "magenta":
		c = color.New(color.FgMagenta)
	default:
		c = color.New(color.FgWhite)
	}

	// 打印带颜色的文本
	_, err := c.Println(text)
	if err != nil {
		return
	}
}

func Timelog(text string, color string) {
	currentTime := time.Now()
	formattedTime := "[eeee] " + currentTime.Format("2006/01/02 15:04:05") + " " // 输出格式：YYYY/MM/DD HH:MM:SS
	logtext := formattedTime + text
	PrintColor(logtext, color)
}

func IpToCIDR(ip string) (string, error) {
	ipa := strings.Split(ip, ".")
	result := strings.Join(ipa[:len(ipa)-1], ".")
	cidr := fmt.Sprintf("%s.0/24", result)
	return cidr, nil
}
