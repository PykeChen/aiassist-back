package main

import (
	"hy_golang_sdk/pkg"
	"log"
	"regexp"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

func main() {
	// 连接到服务器
	c, err := client.DialTLS("imap.qq.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to server.")

	// 登录
	if err := c.Login(pkg.QqMail, pkg.QqPassword); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in.")

	// 选择收件箱
	_, err = c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	// 获取过去两小时的邮件
	criteria := imap.NewSearchCriteria()
	criteria.Since = time.Now().Add(-2 * time.Hour)
	criteria.Header.Set("From", "system@notice.aliyun.com")
	ids, err := c.Search(criteria)
	if err != nil {
		log.Fatal(err)
	}

	// 获取邮件详情
	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchBody}, messages)
	}()

	// 链接正则表达式
	re := regexp.MustCompile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(\\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

	// 读取邮件
	for msg := range messages {
		if msg == nil {
			log.Fatal("Server didn't returned message")
		}

		section, _ := imap.ParseBodySectionName("BODY[]")
		r := msg.GetBody(section)
		if r == nil {
			log.Fatal("Server didn't returned message body")
		}

		// 创建邮件阅读器
		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Fatal(err)
		}

		// 读取邮件正文
		for {
			p, err := mr.NextPart()
			if err != nil {
				log.Fatal(err)
			}

			if p.Header.Get("Content-Type") == "text/plain" {
				b := make([]byte, 1024)
				_, err := p.Body.Read(b)
				if err != nil {
					log.Fatal(err)
				}

				// 查找链接
				match := re.FindString(string(b))
				if match != "" {
					log.Println("Link: ", match)
				}
			}
		}
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
