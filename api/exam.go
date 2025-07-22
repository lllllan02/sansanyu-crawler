package api

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Exam struct {
	Title    string
	Problems []*Problem
}

type Problem struct {
	Type     string
	Content  string
	Options  []string
	Answer   string
	Analysis string
}

func (client *Client) GetExam(url string) (*Exam, error) {
	// 获取试卷详情
	body, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	// 解析 html
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(doc.Find(".p-10").Text())

	var problems []*Problem
	doc.Find("#paper .content-box").Each(func(i int, s *goquery.Selection) {
		problem := &Problem{}

		// 解析标题
		html, _ := s.Find(".title").Find("a").Remove().End().Html()
		problem.Type = parseTitle(html)

		s.Find("li").Each(func(i int, s *goquery.Selection) {
			switch i {

			// 获取题面
			case 0:
				html, _ := s.Html()
				problem.Content = strings.TrimSpace(html)

			// 获取选项
			case 1:
				problem.Options = parseOption(s)

			// 获取答案
			case 3:
				problem.Answer = parseAnswer(problem.Type, s)

			// 获取解析
			case 4:
				html, _ := s.Find(".col-xs-11").Html()
				problem.Analysis = strings.TrimSpace(html)
			}
		})

		problems = append(problems, problem)
	})

	return &Exam{
		Title:    title,
		Problems: problems,
	}, nil
}

func parseTitle(html string) string {
	html = strings.TrimSpace(html)

	left, right := strings.Index(html, "【"), strings.Index(html, "】")
	if left != -1 && right != -1 {
		html = html[left+4 : right-1]
	}

	return fmt.Sprintf("%s", html)
}

func parseOption(s *goquery.Selection) []string {
	var options []string
	s.Find("p").Each(func(i int, s *goquery.Selection) {
		html, _ := s.Html()
		// 去除选项编号，如 "A：" 或 "A."
		if idx := strings.Index(html, "："); idx != -1 {
			html = strings.TrimSpace(html[idx+len("："):])
		} else if idx := strings.Index(html, "."); idx != -1 {
			html = strings.TrimSpace(html[idx+1:])
		}
		options = append(options, html)
	})
	return options
}

func parseAnswer(t string, s *goquery.Selection) string {
	if t == "单选题" || t == "判断题" {
		text := s.Find(".col-xs-11").Text()
		return strings.TrimSpace(text)
	}

	html, _ := s.Find(".col-xs-11").Html()
	return strings.TrimSpace(html)
}
