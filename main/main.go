package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const version = "1.5"

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
	fmt.Printf(usage, argNum)
	fmt.Printf("\n按回车键继续...\n")
	_, _ = fmt.Scanln()
}

func main() {
	fmt.Printf(title)
	fmt.Printf(github)
	fmt.Println()
	const argNum = 4
	if len(os.Args) < argNum+1 {
		help(argNum)
	} else {
		carrierLabel := os.Args[3]
		carrier := url.QueryEscape(map[string]string{
			"校园网":  "",
			"中国移动": "@cmcc",
			"中国联通": "@unicom",
			"中国电信": "@telecom",
		}[carrierLabel])
		id := url.QueryEscape(os.Args[1])
		pwd := url.QueryEscape(os.Args[2])
		sec := os.Args[4]
		interval, err := strconv.Atoi(sec)
		if err != nil {
			fmt.Printf("\n参数错误：无法将参数 %s 解析为秒数\n", sec)
			help(argNum)
		} else {
			for tested, redirect, params := false, false, []string{}; ; redirect, params = false, nil {
				fmt.Println(time.Now().String() + "：")
				u, _ := url.Parse("http://www.baidu.com")
				r := &http.Request{
					Method: http.MethodGet,
					URL:    u,
					Header: map[string][]string{
						"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36 Edg/107.0.1418.42"},
					},
				}
				res, err := (&http.Client{
					Timeout: 2 * time.Second,
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						if strings.Contains(req.URL.String(), "www.baidu.com/") {
							return http.ErrUseLastResponse
						}
						redirect = true
						params = append(params, "wlan_user_ip="+req.URL.Query().Get("wlanuserip"))
						params = append(params, "wlan_user_ipv6="+req.URL.Query().Get("wlanuseripv6"))
						params = append(params, "wlan_user_mac="+strings.ReplaceAll(req.URL.Query().Get("wlanusermac"), "-", ""))
						params = append(params, "wlan_ac_ip="+req.URL.Query().Get("wlanacip"))
						params = append(params, "wlan_ac_name="+req.URL.Query().Get("wlanacname"))
						return http.ErrUseLastResponse
					},
				}).Do(r)
				timeout := err != nil || res.StatusCode == http.StatusBadGateway
				if timeout {
					res, err := http.Get("http://10.0.1.5")
					if err == nil {
						htmlBA, _ := io.ReadAll(res.Body)
						_ = res.Body.Close()
						var html string
						if len(htmlBA) > 0 {
							html = string(htmlBA)
						}
						timeout = strings.Contains(html, "COMWebLoginID_0")
					} else {
						timeout = false
					}
				}
				needLogin := timeout || redirect
				if needLogin {
					var id = id
					if !tested {
						id += carrier
						fmt.Printf("正在尝试用 %s 登录\n", carrierLabel)
					} else {
						fmt.Println("正在尝试用 校园网 登录")
					}
					tested = !tested
					if timeout {
						_, _ = http.Get(
							"http://10.0.1.5/drcom/login?callback=dr1003&DDDDD=" +
								id +
								"&upass=" +
								pwd + "&0MKKey=123456")
					} else if redirect {
						_, _ = http.Get(
							"http://10.0.1.5:801/eportal/portal/login?callback=dr1003&login_method=1&user_account=%2C0%2C" +
								id +
								"&user_password=" +
								pwd +
								"&" +
								strings.Join(params, "&"))
					}
				} else {
					tested = false
					fmt.Println("无需登录")
					time.Sleep(time.Duration(interval) * time.Second)
				}
			}
		}
	}
}
