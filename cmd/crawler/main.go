package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lllllan02/sansanyu-crawler/api"
)

func main() {

	crawler, _ := NewCrawler("exam_currentuser=%2592%25A2%25A4%25A0%25F3%25A9%25AE%25A2%259D%2599%25C5%25DD%25E7%25D9%25DF%25D8%25C2%25D9%259DVk%25E9%25A8%259AS%25B3e%2592%25B6%2583%25C2nf%2599%25A2%25C8%25A9j%25D7%259C%25C7%25A8%2587%25AD%25A8%25CC%25AF%2599%258A%25A9%259Aka%25A8%25D4%25CA%2582%25AC%25A5%2580%258C%25BE%2598ok%258A%25E4%25CB%25EB%25A9%25DD%25D8%25D1%25E0%25C2%259A%25AF%25D9%25B0%259A%2587%25AA%255Bj%2560%25A4%259F%259FW%25A7u%258E%2583y%2593iS%25A3%25E4%25A0%25A9l%25AE%258B%25D6%25DC%25C5%25EB%25DD%25D5%25E4%25DD%25BD%25DD%259E%25A0%2599%25E3%25D7%25DBC%25B4%25AC%2598%2582%2582%2593ia%259E%25A7%259F%25ABn%25AF%25E6; PHPSESSID=31jdop51smjamjmpmup43q0tt0; Hm_lvt_d94f44eb61c2736a4148c9468ea77c46=1753147657; HMACCOUNT=4D5E77D5BB397FC9; Hm_lpvt_d94f44eb61c2736a4148c9468ea77c46=1753151326")

	for _, id := range []int{18, 19, 20} {
		crawler.Subject(id)
	}

}

type Crawler struct {
	client *api.Client
	force  bool
}

func NewCrawler(cookie string) (*Crawler, error) {
	fmt.Println("正在初始化爬虫客户端...")
	client, err := api.NewClient(cookie)
	if err != nil {
		fmt.Printf("初始化爬虫客户端失败: %v\n", err)
		return nil, err
	}
	fmt.Println("爬虫客户端初始化成功")
	return &Crawler{client: client}, nil
}

func (c *Crawler) Force() *Crawler {
	c.force = true
	return c
}

func (c *Crawler) Subject(id int) error {
	fmt.Printf("开始爬取科目 ID: %d\n", id)
	url := fmt.Sprintf("http://www.sansanyu.com/index.php?learn-app-shijuan&subjectid=%d&page=1", id)

	pageNum := 1
	for {
		fmt.Printf("正在爬取第 %d 页...\n", pageNum)
		exams, next, err := c.client.GetSubject(url)
		if err != nil {
			fmt.Printf("爬取科目 %d 第 %d 页失败: %v\n", id, pageNum, err)
			return err
		}
		fmt.Printf("成功获取第 %d 页的试卷列表，共 %d 份试卷\n", pageNum, len(exams))

		for _, exam := range exams {
			fmt.Printf("正在处理试卷: %s\n", exam.Url)
			if err := c.Exam(id, exam.Url); err != nil {
				fmt.Printf("处理试卷失败 [%s]: %v\n", exam.Url, err)
				return err
			}
		}

		if next == "" {
			fmt.Printf("科目 %d 爬取完成\n", id)
			break
		}
		url = next
		pageNum++
	}

	return nil
}

func (c *Crawler) Exam(subject int, link string) error {
	fmt.Printf("开始获取试卷内容: %s\n", link)
	exam, err := c.client.GetExam(link)
	if err != nil {
		fmt.Printf("获取试卷内容失败 [%s]: %v\n", link, err)
		return err
	}

	var id int
	if _, err := fmt.Sscanf(link, "http://www.sansanyu.com/index.php?learn-app-shijuan-detail&examid=%d", &id); err != nil {
		fmt.Printf("解析试卷 ID 失败 [%s]: %v\n", link, err)
		return err
	}

	path := fmt.Sprintf("data/%d/%d.json", subject, id)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		fmt.Printf("创建目录失败 [%s]: %v\n", filepath.Dir(path), err)
		return err
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(exam); err != nil {
		return fmt.Errorf("JSON 序列化失败: %v", err)
	}
	data := buf.Bytes()

	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("保存试卷数据失败 [%s]: %v\n", path, err)
		return err
	}
	fmt.Printf("试卷保存成功: %s\n", path)

	return nil
}
