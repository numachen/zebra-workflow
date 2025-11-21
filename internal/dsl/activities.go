package dsl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"strings"

	"go.temporal.io/sdk/activity"
	"golang.org/x/net/html"

	"zebra-workflow/internal/types"
)

type SampleActivities struct {
}

func (a *SampleActivities) SampleActivity(ctx context.Context, input map[string]string) (map[string]interface{}, error) {
	name := activity.GetInfo(ctx).ActivityType.Name
	fmt.Printf("Run %s with input %v \n", name, input)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 发起GET请求
	url := "https://blog.csdn.net/itopit/article/details/131218450"
	resp, err := client.Get(url)
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// 解析HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// 提取信息
	article := types.ArticleInfo{}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h1" {
			// 查找标题
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "title-article") {
					if n.FirstChild != nil {
						article.Title = n.FirstChild.Data
					}
				}
			}
		}

		if n.Type == html.ElementNode && n.Data == "span" {
			// 查找时间
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "time" {
					if n.FirstChild != nil {
						article.Time = n.FirstChild.Data
					}
				}
			}
		}

		if n.Type == html.ElementNode && n.Data == "div" {
			// 查找内容
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == "content_views" {
					// 提取文本内容
					var extractText func(*html.Node)
					extractText = func(node *html.Node) {
						if node.Type == html.TextNode {
							article.Content += node.Data
						}
						for c := node.FirstChild; c != nil; c = c.NextSibling {
							extractText(c)
						}
					}
					extractText(n)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	// 如果没有找到内容，则使用默认值
	if article.Title == "" {
		article.Title = "未找到标题"
	}
	if article.Time == "" {
		article.Time = "未找到时间"
	}
	if article.Content == "" {
		article.Content = "未找到内容"
	}

	fmt.Printf("Article Info - Title: %s, Time: %s\n", article.Title, article.Time)

	// 构建 map[string]interface{} 类型的结果
	result := map[string]interface{}{
		"title":   article.Title,
		"time":    article.Time,
		"content": article.Content,
	}

	return result, nil
}

func (a *SampleActivities) GetTitle(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {

	// 获取 r1 键的值
	if r1Value, ok := input["r1"]; ok {
		fmt.Println("r1 的值:", r1Value)

		// 类型断言，将 r1Value 转换为字符串
		if r1Str, ok := r1Value.(string); ok {
			// 将字符串解析为 JSON 对象
			var articleData map[string]interface{}
			if err := json.Unmarshal([]byte(r1Str), &articleData); err == nil {
				// 从解析后的数据中获取 title
				if title, ok := articleData["title"]; ok {
					// 类型断言，将 title 转换为 string 类型
					if titleStr, ok := title.(string); ok {
						fmt.Println("文章标题:", titleStr)
						return map[string]interface{}{"标题": titleStr}, nil
					} else {
						fmt.Println("title 字段不是字符串类型")
						return map[string]interface{}{"标题": "未知标题"}, nil
					}
				} else {
					fmt.Println("未找到 title 字段")
				}
			} else {
				fmt.Println("JSON 解析错误:", err)
			}
		} else {
			fmt.Println("r1Value 不是字符串类型")
		}
	} else {
		fmt.Println("未找到 r1 键")
	}

	return map[string]interface{}{"标题": "未知标题"}, nil
}
