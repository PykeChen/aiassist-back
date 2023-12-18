package mail_fetch

import (
	"hy_golang_sdk/pkg"
	"log"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func listMailboxes(c *client.Client, prefix string) {
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List(prefix, "*", mailboxes)
	}()

	for m := range mailboxes {
		log.Println("* " + m.Name)
		listMailboxes(c, m.Name+"/")
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
}

func main3() {
	// 连接到服务器
	c, err := client.DialTLS("imap.qq.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to server.")

	// 登录
	if err := c.Login(pkg.QQ_MAIL, pkg.QQ_PASSWORD); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in.")

	// 获取所有的文件夹
	log.Println("Mailboxes:")
	listMailboxes(c, "")

	log.Println("Done!")
}
