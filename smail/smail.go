package smail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"gobak/config"
	"io/ioutil"
	"log"
	"net/smtp"
)

//todo : BOUNDARY failed: show attachment

//MailSend Sends e-mail with attachments
func MailSend(textMail, subjectMail, fileLocationText, fileNameText string) {
	var body, fileLocation, fileName, from, marker, part1, part2, part3, subject, to, toName string
	var buf bytes.Buffer
	from = config.Current().EmailFrom
	to = config.Current().EmailTo
	toName = "first last"
	marker = "ACUSTOMANDUNIQUEBOUNDARY"
	subject = subjectMail
	body = textMail
	fileLocation = fileLocationText
	fileName = fileNameText

	//part 1 will be the mail headers
	part1 = fmt.Sprintf("From: "+from+" <%s>\r\nTo: %s <%s>\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n--%s", from, toName, to, subject, marker, marker)

	//	part 2 will be the body of the email (text or HTML)
	part2 = fmt.Sprintf("\r\nContent-Type: text/html\r\nContent-Transfer-Encoding:8bit\r\n\r\n%s\r\n--%s", body, marker)

	//read and encode attachment
	content, _ := ioutil.ReadFile(fileLocation)
	encoded := base64.StdEncoding.EncodeToString(content)

	//split the encoded file in lines (doesn't matter, but low enough not to hit a max limit)
	lineMaxLength := 500
	nbrLines := len(encoded) / lineMaxLength

	//append lines to buffer
	for i := 0; i < nbrLines; i++ {
		if _, e := buf.WriteString(encoded[i*lineMaxLength:(i+1)*lineMaxLength] + "\n"); e != nil {
			log.Println(e)
		}
	} //for

	//append last line in buffer
	if _, e := buf.WriteString(encoded[nbrLines*lineMaxLength:]); e != nil {
		log.Println(e)
	}

	//part 3 will be the attachment
	part3 = fmt.Sprintf("\r\nContent-Type: application/csv; name=\"%s\"\r\nContent-Transfer-Encoding:base64\r\nContent-Disposition: attachment; filename=\"%s\"\r\n\r\n%s\r\n--%s--", fileLocation, fileName, buf.String(), marker)

	//send the email
	err := smtp.SendMail(config.Current().SMTPServer, nil, from, []string{to}, []byte(part1+part2+part3))

	//check for SendMail error
	if err != nil {
		log.Fatal(err)
	} //if
}
