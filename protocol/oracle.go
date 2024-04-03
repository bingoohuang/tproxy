package protocol

import (
	"bufio"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/axgle/mahonia"
)

type oracleInterop struct{}

func (m oracleInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	buf := bufio.NewReader(r)

	buffer := make([]byte, 5120)
	for {
		n, err := buf.Read(buffer)
		if err != nil {
			log.Printf("error: %v", err)
			break
		}

		if source == ClientSide {
			r := regexp.MustCompile(`[\s\r\n]+`)
			sql := r.ReplaceAllString(string(buffer[:n]), " ")
			sql = mahonia.NewDecoder("gbk").ConvertString(sql)
			sql = strings.ReplaceAll(sql, "@", "")
			sql = strings.ReplaceAll(sql, "�", "")
			sql = strings.TrimSpace(sql)

			if matches := re.FindAllString(sql, -1); len(matches) > 0 {
				log.Printf("SQL: %s", sql)
			}
		}
	}
}

var re = regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE|CREATE|DROP)\b.*`)
