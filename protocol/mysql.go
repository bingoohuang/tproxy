package protocol

import (
	"bufio"
	"fmt"
	"io"
	"log"
)

type mysqlInterop struct{}

func (m mysqlInterop) Dump(r io.Reader, source string, id int, quiet bool) {

	// only parse client send command
	buf := bufio.NewReader(r)

	buffer := make([]byte, 5120)
	for {
		n, err := buf.Read(buffer)
		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		if n > 4 && source == ClientSide {
			var verboseStr string
			switch buffer[4] {
			case comQuit:
				verboseStr = fmt.Sprintf("Quit: %s\n", "user quit")
			case comInitDB:
				verboseStr = fmt.Sprintf("schema: use %s\n", buffer[5:n])
			case comQuery:
				verboseStr = fmt.Sprintf("Query: %s\n", buffer[5:n])
			case comFieldList:
				verboseStr = fmt.Sprintf("Table columns list: %s\n", buffer[5:n])
			case comCreateDB:
				verboseStr = fmt.Sprintf("CreateDB: %s\n", buffer[5:n])
			case comDropDB:
				verboseStr = fmt.Sprintf("DropDB: %s\n", buffer[5:n])
			case comRefresh:
				verboseStr = fmt.Sprintf("Refresh: %s\n", buffer[5:n])
			case comStmtPrepare:
				verboseStr = fmt.Sprintf("Prepare Query: %s\n", buffer[5:n])
			case comStmtExecute:
				verboseStr = fmt.Sprintf("Prepare Args: %s\n", buffer[5:n])
			case comProcessKill:
				verboseStr = fmt.Sprintf("Kill: kill conntion %s\n", buffer[5:n])
			default:
			}

			if verboseStr != "" {
				log.Print(verboseStr)
			}
		}
	}
}

// read more client-server protocol from http://dev.mysql.com/doc/internals/en/text-protocol.html
const (
	comQuit byte = iota + 1
	comInitDB
	comQuery
	comFieldList
	comCreateDB
	comDropDB
	comRefresh
	comShutdown
	comStatistics
	comProcessInfo
	comConnect
	comProcessKill
	comDebug
	comPing
	comTime
	comDelayedInsert
	comChangeUser
	comBinlogDump
	comTableDump
	comConnectOut
	comRegiserSlave
	comStmtPrepare
	comStmtExecute
	comStmtSendLongData
	comStmtClose
	comStmtReset
	comSetOption
	comStmtFetch
)
