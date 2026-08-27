package tool

import (
	"fmt"
	"github.com/anaskhan96/soup"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"
)

func TestCCNULogin(t *testing.T) {
	t.Skip("手动调试脚本：依赖真实 CCNU 网络与写死账号，默认跳过")
	username := "xxx"
	password := "xxx"
	resp, _ := soup.Get("https://account.ccnu.edu.cn/cas/login?service=http%3A%2F%2Fone.ccnu.edu.cn%2Fcas%2Flogin_portal")
	doc := soup.HTMLParse(resp)
	links1 := doc.Find("body", "id", "cas").FindAll("script")
	js := links1[2].Attrs()["src"][26:]
	links2 := doc.Find("div", "class", "logo").FindAll("input")

	st := links2[2].Attrs()["value"]
	jar, _ := cookiejar.New(&cookiejar.Options{})

	client := &http.Client{
		Jar:     jar,
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://account.ccnu.edu.cn/cas/login;jsessionid=%v?service=http", js) + "%3A%2F%2Fone.ccnu.edu.cn%2Fcas%2Flogin_portal"
	text := fmt.Sprintf("username=%v&password=%v&lt=%v&execution=e1s1&_eventId=submit&submit=", username, password, st) + "%E7%99%BB%E5%BD%95"
	body := strings.NewReader(text)
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Cookie", "JSESSIONID="+js)
	req.Header.Set("Host", "account.ccnu.edu.cn")
	req.Header.Set("Origin", "https://account.ccnu.edu.cn")
	req.Header.Set("Referer", "https://account.ccnu.edu.cn/cas/login?service=http%3A%2F%2Fone.ccnu.edu.cn%2Fcas%2Flogin_portal")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, _ := client.Do(req)
	fmt.Println(len(res.Cookies()))
}

func TestParseCasLogin(t *testing.T) {
	validHTML := `<html><body id="cas">` +
		`<script src="https://account.ccnu.edu.cn/cas/js/a.js"></script>` +
		`<script src="https://account.ccnu.edu.cn/cas/js/b.js"></script>` +
		`<script src="https://account.ccnu.edu.cn/cas/js/c.js"></script>` +
		`<div class="logo"><input/><input/><input value="LT-12345"/></div>` +
		`</body></html>`

	cases := []struct {
		name   string
		html   string
		ok     bool
		wantJS string
		wantST string
	}{
		{"empty", "", false, "", ""},
		{"short", "<html><body></body></html>", false, "", ""},
		{"no-cas", "<html><body><div>hello</div></body></html>", false, "", ""},
		{"few-scripts", `<html><body id="cas"><script src="https://a.com/1.js"></script></body></html>`, false, "", ""},
		{"short-src", `<html><body id="cas"><script src="https://a.com/1.js"></script><script src="https://a.com/2.js"></script><script src="x"></script></body></html>`, false, "", ""},
		{"no-logo", `<html><body id="cas"><script src="https://account.ccnu.edu.cn/cas/js/a.js"></script><script src="https://account.ccnu.edu.cn/cas/js/b.js"></script><script src="https://account.ccnu.edu.cn/cas/js/c.js"></script><div>no logo</div></body></html>`, false, "", ""},
		{"valid", validHTML, true, "n/cas/js/c.js", "LT-12345"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			js, st, ok := parseCasLogin(tc.html)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v (js=%q st=%q)", tc.ok, ok, js, st)
			}
			if tc.ok && (js != tc.wantJS || st != tc.wantST) {
				t.Fatalf("expected js=%q st=%q, got js=%q st=%q", tc.wantJS, tc.wantST, js, st)
			}
		})
	}
}
