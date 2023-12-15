package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const version = "2.5"

const title = "\nGUET校园网自动登录 v" + version + "\n"

const github = "\n代码开源于：https://github.com/TelephoneTan/GUETLANAutoLogin\n"

const usage = "\n使用方法：\n" +
	"\n" +
	"此程序运行需要 %d 个参数：\n" +
	"① 账号\n" +
	"② 密码\n" +
	"③ 运营商（必须是下列值中的一个：校园网，中国移动，中国联通，中国电信）\n" +
	"④ 脚本运行间隔（单位是秒，例如：5 表示每 5 秒测试一次，如果发现掉线则自动登录）\n"

func help(argNum int) {
	log.Printf(usage, argNum)
	log.Printf("\n按回车键继续...\n")
	_, _ = fmt.Scanln()
}

func run(args []string) {
	log.Printf(title)
	log.Printf(github)
	log.Println()
	const argNum = 4
	if len(os.Args) < argNum+1 {
		help(argNum)
	} else {
		carrierLabel := args[3]
		carrier := map[string]string{
			"校园网":  "",
			"中国移动": "@cmcc",
			"中国联通": "@unicom",
			"中国电信": "@telecom",
		}[carrierLabel]
		id := args[1]
		rawPWD := args[2]
		encodedPWD := base64.StdEncoding.EncodeToString([]byte(rawPWD))
		sec := args[4]
		interval, err := strconv.Atoi(sec)
		if err != nil {
			log.Printf("\n参数错误：无法将参数 %s 解析为秒数\n", sec)
			help(argNum)
		} else {
			const us = "http://119.29.29.29"
			u, _ := url.Parse(us)
			r := &http.Request{
				Method: http.MethodGet,
				URL:    u,
				Header: map[string][]string{
					"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36 Edg/107.0.1418.42"},
				},
			}
			client := http.Client{
				Timeout: 2 * time.Second,
			}
			carrierTryTimes := 2
			for tested, needLogin, params := carrierTryTimes, false, map[string]string{}; ; needLogin, params = false, map[string]string{} {
				log.Println(time.Now().String() + "：")
				client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					log.Printf("访问 %s 时被重定向到 %s ，需要登录：", us, req.URL.String())
					needLogin = true
					params["wlan_user_ip"] = req.URL.Query().Get("wlanuserip")
					params["wlan_user_ipv6"] = req.URL.Query().Get("wlanuseripv6")
					params["wlan_user_mac"] = strings.ReplaceAll(req.URL.Query().Get("wlanusermac"), "-", "")
					params["wlan_ac_ip"] = req.URL.Query().Get("wlanacip")
					params["wlan_ac_name"] = req.URL.Query().Get("wlanacname")
					return http.ErrUseLastResponse
				}
				res, err := client.Do(r)
				if err != nil {
					log.Printf("访问 %s 时发生了错误：", us)
					log.Printf("%+#v", err)
					log.Println(err)
				} else {
					log.Printf("访问 %s 时返回状态码：", us)
					log.Println(res.StatusCode)
				}
				networkFailed := err != nil || res.StatusCode == http.StatusBadGateway
				if err == nil {
					_, _ = io.ReadAll(res.Body)
					_ = res.Body.Close()
				}
				if networkFailed {
					const us = "http://10.0.1.5"
					res, err := http.Get(us)
					if err == nil {
						log.Printf("访问 %s 时返回状态码，有网：", us)
						log.Println(res.StatusCode)
						htmlBA, _ := io.ReadAll(res.Body)
						_ = res.Body.Close()
						if len(htmlBA) > 0 {
							html := string(htmlBA)
							if strings.Contains(html, "COMWebLoginID_0") {
								log.Println("校园网内网关异常，需要登录")
								needLogin = true
							} else {
								log.Println("非校园网，无需登录")
							}
						} else {
							log.Println("非校园网，无需登录")
						}
					} else {
						log.Printf("访问 %s 时发生了错误，没网，无需登录：", us)
						log.Printf("%+#v", err)
						log.Println(err)
					}
				}
				if needLogin {
					var id = id
					if tested > 0 {
						tested--
						id += carrier
						log.Printf("正在尝试用 %s 登录\n", carrierLabel)
					} else {
						tested = carrierTryTimes
						log.Println("正在尝试用 校园网 登录")
					}
					wifiLogin := false
					{
						q := url.Values{
							"DDDDD": {id},
							"upass": {rawPWD},
						}
						for k, v := range params {
							q.Set(k, v)
						}
						const us = "http://10.0.1.5/drcom/login"
						// 这里特别注意：参数必须按照特定顺序排列，否则会影响服务器行为
						res, err := http.Get(us + "?callback=dr1003&" + q.Encode() + "&0MKKey=123456")
						if err == nil {
							bs, err := io.ReadAll(res.Body)
							_ = res.Body.Close()
							if err != nil {
								log.Printf("访问 %s 时发生了错误，需要进行无线登录：", us)
								log.Printf("%+#v", err)
								log.Println(err)
								wifiLogin = true
							} else {
								log.Printf("访问 %s 时返回状态码和内容：", us)
								log.Println(res.StatusCode)
								log.Println(string(bs))
								if res.StatusCode == http.StatusNotFound {
									log.Println("非以太网，需要进行无线登录")
									wifiLogin = true
								}
							}
						} else {
							log.Printf("访问 %s 时发生了错误，需要进行无线登录：", us)
							log.Printf("%+#v", err)
							log.Println(err)
							wifiLogin = true
						}
					}
					if wifiLogin {
						q := url.Values{
							"callback":      {"dr1003"},
							"login_method":  {"1"},
							"user_account":  {",0," + id},
							"user_password": {encodedPWD},
						}
						for k, v := range params {
							q.Set(k, v)
						}
						const us = "http://10.0.1.5:801/eportal/portal/login"
						res, err := http.Get(us + "?" + q.Encode())
						if err == nil {
							bs, err := io.ReadAll(res.Body)
							_ = res.Body.Close()
							if err != nil {
								log.Printf("访问 %s 时发生了错误：", us)
								log.Printf("%+#v", err)
								log.Println(err)
							} else {
								log.Printf("访问 %s 时返回状态码和内容：", us)
								log.Println(res.StatusCode)
								log.Println(string(bs))
							}
						} else {
							log.Printf("访问 %s 时发生了错误：", us)
							log.Printf("%+#v", err)
							log.Println(err)
						}
					}
				} else {
					tested = carrierTryTimes
					log.Println("无需登录")
					time.Sleep(time.Duration(interval) * time.Second)
				}
			}
		}
	}
}

func main() {
	run(os.Args)
}
