package tool

import (
	"fmt"
	"github.com/anaskhan96/soup"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

func CCNULogin(studentID string, password string) bool {
	htmlBody, err := soup.Get("https://account.ccnu.edu.cn/cas/login?service=http%3A%2F%2Fone.ccnu.edu.cn%2Fcas%2Flogin_portal")
	if err != nil {
		return false
	}
	js, st, ok := parseCasLogin(htmlBody)
	if !ok {
		return false
	}
	jar, _ := cookiejar.New(&cookiejar.Options{})

	client := &http.Client{
		Jar:     jar,
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://account.ccnu.edu.cn/cas/login;jsessionid=%v?service=http", js) + "%3A%2F%2Fone.ccnu.edu.cn%2Fcas%2Flogin_portal"
	text := fmt.Sprintf("username=%v&password=%v&lt=%v&execution=e1s1&_eventId=submit&submit=", studentID, password, st) + "%E7%99%BB%E5%BD%95"
	body := strings.NewReader(text)
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Cookie", "JSESSIONID="+js)
	req.Header.Set("Host", "account.ccnu.edu.cn")
	req.Header.Set("Origin", "https://account.ccnu.edu.cn")
	req.Header.Set("Referer", "https://account.ccnu.edu.cn/cas/login?service=http%3A%2F%2Fone.ccnu.edu.cn%2Fcas%2Flogin_portal")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return false
	}

	return len(resp.Cookies()) != 0
}

// parseCasLogin 从 CCNU CAS 登录页解析 jsessionid（js）与 lt 参数（st）。
// 页面结构不匹配或 HTML 缺失时返回 ok=false，避免对 soup 结果越界访问。
func parseCasLogin(html string) (js, st string, ok bool) {
	doc := soup.HTMLParse(html)
	casBody := doc.Find("body", "id", "cas")
	if casBody.Pointer == nil {
		return "", "", false
	}
	links1 := casBody.FindAll("script")
	if len(links1) < 3 {
		return "", "", false
	}
	src := links1[2].Attrs()["src"]
	if len(src) < 27 {
		return "", "", false
	}
	js = src[26:]
	logo := doc.Find("div", "class", "logo")
	if logo.Pointer == nil {
		return "", "", false
	}
	links2 := logo.FindAll("input")
	if len(links2) < 3 {
		return "", "", false
	}
	st = links2[2].Attrs()["value"]
	return js, st, true
}
