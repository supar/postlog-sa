package main

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"postlog-sa/filter"
	"reflect"
	"time"
)

type StmtMap struct {
	stmt   *sql.Stmt
	params []string
}

func (this *StmtMap) Call(fn filter.ThreadFace) (err error) {
	var (
		call reflect.Value
		res  []reflect.Value
		fn_c = reflect.ValueOf(fn)
		args = make([]interface{}, 0)
	)

	for _, f := range this.params {
		call = fn_c.MethodByName(f)
		res = call.Call(nil)

		if l := len(res); l > 0 {
			v := res[0].Interface()

			switch v.(type) {
			case time.Time:
				args = append(args, v.(time.Time).Format("2006-01-02 15:04:05"))
			default:
				args = append(args, v)
			}
		}
	}

	_, err = this.stmt.Exec(args...)
	return
}

func NewStmt(db *sql.DB, query string) (stmt *StmtMap, err error) {
	var (
		buffer  *bytes.Buffer
		runes   []rune
		fn_name string
	)

	buffer, runes, err = parseStmtQuery(query)
	if err != nil {
		return nil, err
	}

	stmt = &StmtMap{
		params: make([]string, 0),
	}

	if stmt.stmt, err = db.Prepare(buffer.String()); err != nil {
		return nil, err
	}

	for _, r := range runes {
		fn_name = ""

		switch r {
		// c
		case 99:
			fn_name = "GetFromIp"
		// f
		case 102:
			fn_name = "GetFrom"
		// i
		case 105:
			fn_name = "GetId"
		// m
		case 109:
			fn_name = "GetMessageId"
		// s
		case 115:
			fn_name = "GetSpamScore"
		// t
		case 116:
			fn_name = "GetTime"
		}

		if fn_name != "" {
			stmt.params = append(stmt.params, fn_name)
		}
	}
	return
}

/**
 * GetId - ?i
 * GetMessageId - ?m
 * GetFrom - ?f
 * GetFromIp - ?c
 * GetTime - ?t
 * GetSpamScore - ?s
 */
func parseStmtQuery(query string) (buffer *bytes.Buffer, runes []rune, err error) {
	var (
		reader = bytes.NewReader([]byte(query))

		pos, m_pos int
		char       rune
	)

	if reader.Len() == 0 {
		return nil, nil, errors.New("Empty reader")
	}

	buffer = bytes.NewBuffer(make([]byte, 0))
	runes = make([]rune, 0)
	pos = -1

	for {
		// Read rune from buffer
		char, pos, err = reader.ReadRune()
		if err != nil {
			// End of buffer
			if err == io.EOF {
				err = nil

				break
			}

			return nil, nil, err
		}

		switch true {
		// Reader at ? rune
		case (char == 63):
			m_pos = pos

		// Previous rune was ?
		case (m_pos > -1):
			m_pos = -1

			switch char {
			case 99, 102, 105, 109, 115, 116:
				runes = append(runes, char)
				continue
			}
		}

		buffer.WriteRune(char)
	}

	return
}
