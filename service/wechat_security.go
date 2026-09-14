package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	accessToken string
	tokenExpiry time.Time
	tokenMutex  sync.Mutex
)

func getAccessToken() (string, error) {
	appID := os.Getenv("WECHAT_APPID")
	appSecret := os.Getenv("WECHAT_APPSECRET")

	if appID == "" || appSecret == "" {
		// 未配置 AppID/Secret，暂时放行（方便本地测试）
		return "", nil
	}

	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	if time.Now().Before(tokenExpiry) && accessToken != "" {
		return accessToken, nil
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if res.ErrCode != 0 {
		return "", fmt.Errorf("failed to get access token: %s", res.ErrMsg)
	}

	accessToken = res.AccessToken
	tokenExpiry = time.Now().Add(time.Duration(res.ExpiresIn-60) * time.Second)

	return accessToken, nil
}

// CheckContentSecurity 检查文本内容是否安全（敏感词过滤）
func CheckContentSecurity(content string) error {
	token, err := getAccessToken()
	if err != nil {
		return err // 获取 token 失败
	}

	if token == "" {
		// 未配置环境变量，跳过检查
		return nil
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/msg_sec_check?access_token=%s", token)

	reqBody := map[string]string{
		"content": content,
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var res struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}

	// 87014 表示内容含有违法违规内容
	if res.ErrCode == 87014 {
		return fmt.Errorf("内容包含违规敏感词汇，请修改后重试")
	} else if res.ErrCode != 0 {
		return fmt.Errorf("安全检查接口调用失败: %s", res.ErrMsg)
	}

	return nil
}
