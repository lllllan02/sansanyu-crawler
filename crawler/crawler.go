package crawler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lllllan02/sansanyu-crawler/api"
)

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
