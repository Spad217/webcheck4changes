package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"encoding/hex"

	"github.com/PuerkitoBio/goquery"
)

func getUrlData(url string, data interface{}) (string, error) {
	funcMap := template.FuncMap{
		"time": func() int64 { return time.Now().Unix() },
	}

	tmpl, err := template.New("titleTest").Funcs(funcMap).Parse(url)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, data)

	return b.String(), err
}

func getUrl(url string) (string, error) {
	return getUrlData(url, struct{}{})
}

func Text2hash(reader io.ReadCloser) (string, error) {
	defer reader.Close()
	hash := md5.New()
	b := make([]byte, 512)
	for {
		n, err := reader.Read(b)
		hash.Write(b[:n])
		if err != nil && err != io.EOF {
			return "", err
		}
		if err == io.EOF {
			break
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func getNodeText(body io.ReadCloser, selector string) (string, error) {
	defer body.Close()
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return "", err
	}
	return doc.Find(selector).Html()
}

func Url2Reader(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	return resp.Body, err
}

func (setting Setting) LoadLink(row *Url) error {
	url, err := getUrl(row.Url)
	if err != nil {
		return err
	}

	reader, err := Url2Reader(url)
	if err != nil {
		return err
	}

	var text string
	if row.Selector != "" {
		text, err = getNodeText(reader, row.Selector)
		reader = io.NopCloser(strings.NewReader(text))
		if err != nil {
			return err
		}
	}

	hash, err := Text2hash(reader)
	if err != nil {
		return err
	}

	if hash != row.Hash {
		err := setting.Trigger(*row, url, text)
		if err != nil {
			return err
		}
		row.Hash = hash
		row.Date = time.Now()
	}
	return nil
}

func (setting Setting) Trigger(row Url, link string, text string) error {
	fmt.Println(time.Now().UTC(), row.Alias)
	message := fmt.Sprintf("[%s](%s)", strings.ReplaceAll(row.Alias, ".", "\\."), url.QueryEscape(link))
	if !row.Disable_Message {
		message = message + fmt.Sprintf(": %s", text)
	}
	url_telegram := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?text=%s&chat_id=%s&parse_mode=MarkdownV2&disable_web_page_preview=1",
		setting.Telegram.Token, message, setting.Telegram.ChatId)
	fmt.Println(url_telegram)
	resp, err := http.Get(url_telegram)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	fmt.Println(buf.String())
	return nil
}
