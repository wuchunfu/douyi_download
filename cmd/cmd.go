package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/nanlei2000/douyin_download/pkg/douyin"
	"github.com/urfave/cli/v2"
)

const MAX_CONCURRENT_NUM = 5

func main() {
	var verbose bool
	var path string
	var downloadUserPost bool

	app := &cli.App{
		Name:  "dydl",
		Usage: "下载抖音视频",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "v",
				Aliases:     []string{"verbose"},
				Value:       false,
				Usage:       "切换 verbose 模式",
				Destination: &verbose,
			},
			&cli.BoolFlag{
				Name:        "up",
				Aliases:     []string{"user_post"},
				Value:       false,
				Usage:       "下载所有发布的视频",
				Destination: &downloadUserPost,
			},
			&cli.StringFlag{
				Name:        "p",
				Aliases:     []string{"path"},
				Value:       "./video/",
				Usage:       "文件下载路径",
				Destination: &path,
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return fmt.Errorf("url is required")
			}

			// https://www.douyin.com/user/MS4wLjABAAAAGTcyBs_MF1p2vK0QB2pUK_N3-huudY2UtA9Shw0o5N8
			if downloadUserPost {
				userLink := c.Args().Get(0)
				parts := strings.Split(userLink, "/")
				secUid := parts[len(parts)-1]
				idList, err := douyin.GetAllVideoIDList(secUid)

				if err != nil {
					return err
				}

				c := make(chan struct{}, MAX_CONCURRENT_NUM)
				defer close(c)

				var wg sync.WaitGroup
				for _, id := range idList {
					wg.Add(1)
					go func(id string) {
						c <- struct{}{}
						defer func() {
							wg.Done()
							<-c
						}()

						dy := douyin.NewDouYin()
						dy.IsDebug(verbose)
						video, err := dy.Get(douyin.Source{
							Type:    douyin.SourceType_VideoID,
							Content: id,
						})
						if err != nil {
							fmt.Printf("get video info failed, id: %s, err: %s", id, err)
						}
						_, err = video.Download(path)
						if err != nil {
							fmt.Printf("download video failed, id: %s, err: %s", id, err)
						}
					}(id)
				}
				wg.Wait()

				return nil
			}

			// use shardContent
			shareContent := ""
			if c.NArg() > 0 {
				shareContent = strings.Join(c.Args().Slice(), "")
			} else {
				return fmt.Errorf("url is required")
			}

			dy := douyin.NewDouYin()
			dy.IsDebug(verbose)
			video, err := dy.Get(douyin.Source{
				Type:    douyin.SourceType_ShardContent,
				Content: shareContent,
			})
			if err != nil {
				return err
			}
			_, err = video.Download(path)
			if err != nil {
				return err
			}

			return nil
		},
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}