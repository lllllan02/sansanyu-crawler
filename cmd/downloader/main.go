package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

const (
	baseURL = "http://www.sansanyu.com"
	imgDir  = "image"
)

// 下载图片并返回新的本地路径
func downloadImage(url string, bar *progressbar.ProgressBar) (string, error) {
	// 确保URL是完整的
	if !strings.HasPrefix(url, "http") {
		url = baseURL + "/" + strings.TrimPrefix(url, "/")
	}

	// 创建HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("下载图片失败: %v", err)
	}
	defer resp.Body.Close()

	// 从URL中提取文件名
	filename := filepath.Base(url)
	localPath := filepath.Join(imgDir, filename)

	// 确保image目录存在
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建本地文件
	out, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	// 使用进度条包装reader
	reader := progressbar.NewReader(resp.Body, bar)

	// 复制内容
	_, err = io.Copy(out, &reader)
	if err != nil {
		return "", fmt.Errorf("保存文件失败: %v", err)
	}

	return localPath, nil
}

// 处理单个JSON文件
func processJSONFile(path string, totalFiles int, currentFile int) error {
	fmt.Printf("正在处理文件: %s\n", path)

	// 读取文件
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 解析 JSON 数据
	var jsonData map[string]interface{}
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return fmt.Errorf("解析 JSON 失败 %s: %v", path, err)
	}

	// 获取 Problems 数组
	problems, ok := jsonData["Problems"].([]interface{})
	if !ok {
		return fmt.Errorf("JSON 格式错误：未找到 Problems 数组")
	}

	// 查找所有图片URL
	re := regexp.MustCompile(`src=\\?"(files/attach/files/content/[^"]*?\.(png|jpg|jpeg|gif))\\?"`)

	var imageURLs []string
	for _, problem := range problems {
		if p, ok := problem.(map[string]interface{}); ok {
			if content, ok := p["Content"].(string); ok {
				matches := re.FindAllStringSubmatch(content, -1)
				for _, match := range matches {
					imageURLs = append(imageURLs, match[1])
				}
			}
		}
	}

	fmt.Printf("在文件 %s 中找到 %d 个图片链接\n", path, len(imageURLs))
	if len(imageURLs) == 0 {
		// 输出一个示例内容用于调试
		if len(problems) > 0 {
			if p, ok := problems[0].(map[string]interface{}); ok {
				if content, ok := p["Content"].(string); ok {
					fmt.Printf("示例内容：\n%s\n", content)
				}
			}
		}
		return nil
	}

	// 创建文件处理进度条
	bar := progressbar.NewOptions(len(imageURLs),
		progressbar.OptionSetDescription(fmt.Sprintf("[%d/%d] 处理文件 %s", currentFile, totalFiles, filepath.Base(path))),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(30),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Printf("\n")
		}),
	)

	// 处理每个图片
	for _, oldURL := range imageURLs {
		fmt.Printf("处理图片URL: %s\n", oldURL)

		// 下载图片
		newPath, err := downloadImage(oldURL, bar)
		if err != nil {
			fmt.Printf("\n警告: 下载图片失败 %s: %v\n", oldURL, err)
			continue
		}

		// 在所有 Problems 中替换图片URL
		for _, problem := range problems {
			if p, ok := problem.(map[string]interface{}); ok {
				if content, ok := p["Content"].(string); ok {
					// 处理转义的情况
					oldPattern := fmt.Sprintf(`src=\\?"%s\\?"`, regexp.QuoteMeta(oldURL))
					newPattern := fmt.Sprintf(`src="%s"`, newPath)
					p["Content"] = regexp.MustCompile(oldPattern).ReplaceAllString(content, newPattern)
				}
			}
		}

		bar.Add(1)
	}

	// 使用指定的编码方式
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jsonData); err != nil {
		return fmt.Errorf("序列化 JSON 失败 %s: %v", path, err)
	}

	// 写入文件
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// 获取需要处理的JSON文件总数
func countJSONFiles(dir string) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".json") {
			count++
		}
		return nil
	})
	return count, err
}

// 遍历目录处理所有JSON文件
func walkDir(dir string) error {
	fmt.Printf("开始遍历目录: %s\n", dir)

	// 获取总文件数
	totalFiles, err := countJSONFiles(dir)
	if err != nil {
		return fmt.Errorf("统计文件数失败: %v", err)
	}

	fmt.Printf("找到 %d 个 JSON 文件\n", totalFiles)

	currentFile := 0
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理JSON文件
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".json") {
			currentFile++
			if err := processJSONFile(path, totalFiles, currentFile); err != nil {
				fmt.Printf("\n处理文件失败 %s: %v\n", path, err)
			}
		}

		return nil
	})
}

func main() {
	fmt.Println("开始处理文件...")

	// 确保image目录存在
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		fmt.Printf("创建image目录失败: %v\n", err)
		return
	}

	// 开始处理data目录
	if err := walkDir("data"); err != nil {
		fmt.Printf("\n处理目录失败: %v\n", err)
		return
	}

	fmt.Println("\n处理完成!")
}
