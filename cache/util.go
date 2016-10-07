package cache

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ContainsBySlice(md5S []string, s string) bool {
	for _, v := range md5S {
		if v == s {
			return true
		}
	}
	return false
}

func RemoveFromSlice(s []string, e string) []string {
	indexS := []int{}
	for i, v := range s {
		if v == e {
			indexS = append(indexS, i)
		}
	}
	for _, v := range indexS {
		s = append(s[:v], s[v+1:]...)
	}
	return s
}

func EncodeUrl(url string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(string(url)))
	cipherEncode := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherEncode)
}

func ConcatCss(resp *http.Response, hostUrl string) ([]byte, error) {
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}
	cssUrls := []string{}
	doc.Find("link[rel=stylesheet]").Each(func(i int, node *goquery.Selection) {
		cssUrl, _ := node.Attr("href")
		if !strings.HasSuffix(hostUrl, "/") {
			hostUrl = hostUrl + "/"
		}
		if !strings.HasPrefix(cssUrl, "http://") {
			cssUrl = hostUrl + cssUrl
		}
		if cssUrl != "" && strings.HasPrefix(cssUrl, hostUrl) {
			cssUrls = append(cssUrls, cssUrl)
			node.Remove()
		}
	})
	cssAll := ""
	for _, v := range cssUrls {
		resp, err := http.Get(v)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		cssAll = cssAll + "\r\n" + string(body)
	}
	doc.Find("title").AfterHtml("\r\n<style>" + cssAll + "</style>\r\n")
	html, _ := doc.Html()
	re, _ := regexp.Compile("\\/\\*.*\\*\\/")
	html = re.ReplaceAllString(html, "")
	return []byte(html), nil
}
