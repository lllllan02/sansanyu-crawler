package api

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ExamLink struct {
	Title string
	Url   string
}

func (client *Client) GetSubject(url string) (exams []ExamLink, next string, err error) {
	// 获取试卷列表
	body, err := client.Get(url)
	if err != nil {
		return nil, "", err
	}

	// 解析 html
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, "", err
	}

	// 查找所有试卷链接
	doc.Find(".list-unstyled .list-group-item a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			exams = append(exams, ExamLink{
				Title: s.Text(),
				Url:   fmt.Sprintf("http://www.sansanyu.com/%s", href),
			})
		}
	})

	// 查找下一页链接
	doc.Find(".pagination li a").Each(func(i int, s *goquery.Selection) {
		// 找到当前页的class="current"
		if s.HasClass("current") {
			// 获取下一个兄弟元素的链接
			if n := s.Parent().Next().Find("a"); n.Length() > 0 {
				if href, exists := n.Attr("href"); exists {
					next = fmt.Sprintf("http://www.sansanyu.com/%s", href)
				}
			}
		}
	})

	return exams, next, nil
}
