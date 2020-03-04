package smail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/smtp"
)

//todo : BOUNDARY failed: show attachment

type MailApp struct {
	smtpServerUrl string
	log           *zap.Logger
	emailTo       string
	emailFrom     string
}

func NewMailApp(smtpServerUrl string, log *zap.Logger, emailTo string, emailFrom string) *MailApp {
	return &MailApp{
		smtpServerUrl: smtpServerUrl,
		log:           log,
		emailTo:       emailTo,
		emailFrom:     emailFrom,
	}
}

//MailSend Sends e-mail with attachments
func (c *MailApp) MailSend(textMail, subjectMail, fileLocationText, fileNameText string) {
	var body, fileLocation, fileName, marker, part1, part2, part3, subject, toName string
	var buf bytes.Buffer
	toName = "first last"
	marker = "ACUSTOMANDUNIQUEBOUNDARY"
	subject = subjectMail
	body = textMail
	fileLocation = fileLocationText
	fileName = fileNameText

	//part 1 will be the mail headers
	part1 = fmt.Sprintf("From: "+c.emailFrom+" <%s>\r\nTo: %s <%s>\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n--%s", c.emailFrom, toName, c.emailTo, subject, marker, marker)

	//	part 2 will be the body of the email (text or HTML)
	part2 = fmt.Sprintf("\r\nContent-Type: text/html\r\nContent-Transfer-Encoding:8bit\r\n\r\n%s\r\n--%s", body, marker)

	//read and encode attachment
	content, _ := ioutil.ReadFile(fileLocation)
	encoded := base64.StdEncoding.EncodeToString(content)

	//split the encoded file in lines (doesn't matter, but low enough not to hit a max limit)
	lineMaxLength := 500
	nbrLines := len(encoded) / lineMaxLength

	for i := 0; i < nbrLines; i++ {
		if _, e := buf.WriteString(encoded[i*lineMaxLength:(i+1)*lineMaxLength] + "\n"); e != nil {
			log.Println(e)
		}
	}

	if _, e := buf.WriteString(encoded[nbrLines*lineMaxLength:]); e != nil {
		c.log.Info(e.Error())
	}

	part3 = fmt.Sprintf("\r\nContent-Type: application/csv; name=\"%s\"\r\nContent-Transfer-Encoding:base64\r\nContent-Disposition: attachment; filename=\"%s\"\r\n\r\n%s\r\n--%s--", fileLocation, fileName, buf.String(), marker)

	err := smtp.SendMail(c.smtpServerUrl, nil, c.emailFrom, []string{c.emailTo}, []byte(part1+part2+part3))
	if err != nil {
		c.log.Fatal(err.Error())
	}
}
