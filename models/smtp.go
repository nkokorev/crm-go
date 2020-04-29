package models

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/mail"
	"strings"
)

type Email struct {
	Message mail.Message // has header and body
}

func TestSend() error {

	var email = new(Email)

	err := email.LoadBodyFromTemplate("files/example.html", nil)
	if err != nil {
		return err
	}

	fmt.Println(email.GetBodyBase64String())
	return err
}

func (email *Email) LoadBodyFromTemplate(addr string, T interface{}) error {
	tpl, err := template.ParseFiles(addr)
	if err != nil {
		return err
	}

	var buf = new(bytes.Buffer)
	err = tpl.Execute(buf, T)

	msg, err := mail.ReadMessage(buf)
	if err != nil { return err }

	email.Message = *msg

	return nil
}

func (email Email) GetBodyByte() ([]byte, error) {
	return ioutil.ReadAll(email.Message.Body)
}

// Возвращает Body
func (email Email) GetBodyString() (string, error) {
	body, err := ioutil.ReadAll(email.Message.Body)
	if err != nil { return "", err}

	return string(body), nil
}

// Возвращает экранированное Body
func (email Email) GetBodyStringEscaped() (string, error) {
	body, err := email.GetBodyString()
	if err != nil {
		return "",err
	}
	return template.HTMLEscapeString(body), nil
}

type linesplitter struct {
	len   int
	count int
	sep   []byte
	w     io.Writer
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func (ls *linesplitter) Close() (err error) {
	return nil
}
// возвращает body в base64 по 76 символов в длину
func (email Email) GetBodyBase64Byte() ([]byte, error) {

	bufR := new(bytes.Buffer)

	lsWriter := &linesplitter{len: 76, count: 0, sep: []byte("\r\n"), w: bufR}
	wrt := base64.NewEncoder(base64.StdEncoding, lsWriter)
	body, err := email.GetBodyStringEscaped()
	if err != nil {return nil, err}
	_, err = io.Copy(wrt, strings.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}

	return bufR.Bytes(), nil

}

func (ls *linesplitter) Write(in []byte) (n int, err error) {
	writtenThisCall := 0
	readPos := 0
	// Leading chunk size is limited by: how much input there is; defined split length; and
	// any residual from last time
	chunkSize := min(len(in), ls.len-ls.count)
	// Pass on chunk(s)
	for {
		ls.w.Write(in[readPos:(readPos + chunkSize)])
		readPos += chunkSize // Skip forward ready for next chunk
		ls.count += chunkSize
		writtenThisCall += chunkSize

		// if we have completed a chunk, emit a separator
		if ls.count >= ls.len {
			ls.w.Write(ls.sep)
			writtenThisCall += len(ls.sep)
			ls.count = 0
		}
		inToGo := len(in) - readPos
		if inToGo <= 0 {
			break // reached end of input data
		}
		// Determine size of the NEXT chunk
		chunkSize = min(inToGo, ls.len)
	}
	return writtenThisCall, nil
}

func (email Email) GetBodyBase64String() (string, error) {

	bodyByte, err := email.GetBodyBase64Byte()
	if err != nil { return "", err}

	return string(bodyByte), nil

}