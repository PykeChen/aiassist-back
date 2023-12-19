package mail_fetch

import (
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"hy_golang_sdk/pkg"
	"hy_golang_sdk/pkg/plog"
	"io"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"time"
)

// 链接正则表达式
var re = regexp.MustCompile(`违规文件链接：.*>(http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(\\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+)<`)

func FetchOssViolationUrl(linkChan chan string) {
	defer func() {
		close(linkChan)
	}()
	// 连接到服务器
	c, err := client.DialTLS("imap.qq.com:993", nil)
	if err != nil {
		plog.Warnf("client dial tls error:%v", err)
	}
	log.Println("Connected to server.")

	// 登录
	if err := c.Login(pkg.QqMail, pkg.QqPassword); err != nil {
		log.Fatal(err)
	}
	plog.Infof("%v:Logged in", time.Now().Format("2006-01-02 15:04:05"))

	// 选择阿里云文件夹
	_, err = c.Select("其他文件夹/阿里云", false)
	if err != nil {
		log.Fatal(err)
	}
	// 搜索条件实例对象
	criteria := imap.NewSearchCriteria()
	criteria.Since = time.Now().Add(-1 * time.Hour)
	//criteria.Header.Set("From", "system@notice.aliyun.com")
	// ALL是默认条件
	// See RFC 3501 section 6.4.4 for a list of searching criteria.
	criteria.WithoutFlags = []string{"\\Seen"}
	ids, _ := c.Search(criteria)
	var s imap.BodySectionName

	for {
		if len(ids) == 0 {
			break
		}
		id := pop(&ids)

		seqset := new(imap.SeqSet)
		seqset.AddNum(id)
		chanMessage := make(chan *imap.Message, 1)
		go func() {
			// 第一次fetch, 只抓取邮件头，邮件标志，邮件大小等信息，执行速度快
			if err = c.Fetch(seqset,
				[]imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchRFC822Size},
				chanMessage); err != nil {
				// 【实践经验】这里遇到过的err信息是：ENVELOPE doesn't contain 10 fields
				// 原因是对方发送的邮件格式不规范，解析失败
				// 相关的issue: https://github.com/emersion/go-imap/issues/143
				plog.Warnf("seq_set:%v, err:%v", seqset, err)
				chanMessage <- nil
			}
		}()

		message := <-chanMessage
		if message == nil {
			plog.Warnf("Server didn't returned message")
			continue
		}

		msgDate := message.Envelope.Date.Format("2006-01-02 15:04:05")
		plog.Infof("time:%v, %v: %v bytes, flags=%v \n", msgDate, message.SeqNum, message.Size, message.Flags)

		if strings.HasPrefix(message.Envelope.Subject, "OSS违规文件冻结处理通知") {
			chanMsg := make(chan *imap.Message, 1)
			go func() {
				// 这里是第二次fetch, 获取邮件MIME内容
				if err = c.Fetch(seqset,
					[]imap.FetchItem{imap.FetchRFC822},
					chanMsg); err != nil {
					log.Println(seqset, err)
				}
			}()

			msg := <-chanMsg
			if msg == nil {
				plog.Errorf("Server didn't returned message")
				continue
			}

			section := &s
			r := msg.GetBody(section)
			if r == nil {
				plog.Errorf("Server didn't returned message body")
				continue
			}

			// Create a new mail reader
			// 创建邮件阅读器
			mr, err := mail.CreateReader(r)
			if err != nil {
				plog.Errorf("create reader err:%v", err)
				continue
			}

			// Process each message's part
			// 处理消息体的每个part
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					plog.Errorf("read part err:%v", err)
					break
				}

				if p.Header.Get("Content-Type") == "text/html;charset=utf-8" {
					b, _ := ioutil.ReadAll(p.Body)
					// 查找链接
					//log.Println("前面Got text: ", string(b))
					match := re.FindStringSubmatch(string(b))
					if len(match) > 1 {
						plog.Infof("Find mail, Date:%v, Link: %v ", msgDate, match[1])
						if linkChan != nil {
							linkChan <- match[1]
						}
					}
				}
				//switch h := p.Header.(type) {
				//case *mail.InlineHeader:
				//	// This is the message's text (can be plain-text or HTML)
				//	// 获取正文内容, text或者html
				//	b, _ := ioutil.ReadAll(p.Body)
				//	log.Println("Got text: ", string(b))
				//case *mail.AttachmentHeader:
				//	// This is an attachment
				//	// 下载附件
				//	filename, err := h.Filename()
				//	log.Printf("Got filename: %v, err:%v\n", filename, err)
				//}
			}

			// 标记为已读
			seqset := new(imap.SeqSet)
			seqset.AddNum(msg.SeqNum)
			item := imap.FormatFlagsOp(imap.AddFlags, true)
			flags := []interface{}{imap.SeenFlag}
			if err := c.Store(seqset, item, flags, nil); err != nil {
				plog.Errorf("mark read err:%v", err)
				continue
			}
		}
	}
}

func pop(list *[]uint32) uint32 {
	length := len(*list)
	lastEle := (*list)[length-1]
	*list = (*list)[:length-1]
	return lastEle
}

func oldFunction() {
	// 连接到服务器
	c, err := client.DialTLS("imap.qq.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to server.")

	// 登录
	if err := c.Login("xxx", "xxxx"); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in.")

	// 选择阿里云文件夹
	_, err = c.Select("其他文件夹/阿里云", false)
	if err != nil {
		log.Fatal(err)
	}

	// 获取未读邮件
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{"\\Seen"}

	ids, err := c.Search(criteria)
	if err != nil {
		log.Fatal(err)
	}

	// 获取邮件详情
	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	var section imap.BodySectionName
	go func() {
		//done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchBody}, messages)
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchBody, imap.FetchEnvelope, imap.FetchRFC822Size, imap.FetchRFC822}, messages)
	}()

	// 读取邮件
	for msg := range messages {
		if msg == nil {
			log.Fatal("Server didn't returned message")
		}
		// 读取邮件
		for msg := range messages {
			if msg == nil {
				plog.Error("Server didn't returned message")
			}
			if msg.Envelope.Subject == "OSS违规文件冻结处理通知" {
				r := msg.GetBody(&section)
				if r == nil {
					plog.Error("Server didn't returned message body")
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
						b, _ := ioutil.ReadAll(p.Body)
						// 查找链接
						match := re.FindStringSubmatch(string(b))
						if len(match) > 1 {
							plog.Error("Link: ", match[1])
						}
					}
				}
			}
		}

		// 标记为已读
		seqset := new(imap.SeqSet)
		seqset.AddNum(msg.SeqNum)
		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []interface{}{imap.SeenFlag}
		if err := c.Store(seqset, item, flags, nil); err != nil {
			log.Fatal(err)
		}
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
