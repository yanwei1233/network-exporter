package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
	Status      string            `json:"status"`
}

type AlertRequest struct {
	Alerts            []Alert           `json:"alerts"`
	Status            string            `json:"status"`
	GroupKey          string            `json:"groupKey"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
}

type WeChatMessage struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}

func generateWeChatMessage(alertRequest AlertRequest, env string) string {
	var message string

	if len(alertRequest.Alerts) > 0 {
		for _, alert := range alertRequest.Alerts {
			if alert.Status == "firing" {
				firingTime := alert.StartsAt.Add(8 * time.Hour)

				message += "🔥<font color=\"warning\">**【告警通知】**</font>🔥\n"
				message += "**告警程序:**" + env + "\n"
				message += "**告警级别:** " + alert.Labels["severity"] + "\n"
				message += "**告警名称:** " + alert.Labels["alertname"] + "\n"
				message += "**告警状态:** " + alert.Status + "\n"
				message += "**告警主机:** " + alert.Labels["instance"] + "\n"
				message += "**告警主题:** " + alert.Annotations["summary"] + "\n"
				message += "**告警详情:** " + alert.Annotations["description"] + "\n"
				message += "**触发时间:** <font color=\"warning\">" + firingTime.Format("2006-01-02 15:04:05") + "</font>\n"
				message += "========= =end= ========\n\n"
			}
		}
	}

	if len(alertRequest.Alerts) > 0 {
		for _, alert := range alertRequest.Alerts {
			if alert.Status == "resolved" {
				firingTime := alert.StartsAt.Add(8 * time.Hour)
				resolveTime := alert.EndsAt.Add(8 * time.Hour)

				message += "✅<font color=\"info\">**【告警恢复】**</font>✅\n"
				message += "**告警程序:**" + env + "\n"
				message += "**告警级别:** " + alert.Labels["severity"] + "\n"
				message += "**告警名称:** " + alert.Labels["alertname"] + "\n"
				message += "**告警状态:** " + alert.Status + "\n"
				message += "**告警主机:** " + alert.Labels["instance"] + "\n"
				message += "**告警主题:** " + alert.Annotations["summary"] + "\n"
				message += "**告警详情:** " + alert.Annotations["description"] + "\n"
				message += "**触发时间:** <font color=\"warning\">" + firingTime.Format("2006-01-02 15:04:05") + "</font>\n"
				message += "**恢复时间:** <font color=\"info\">" + resolveTime.Format("2006-01-02 15:04:05") + "</font>\n"
				message += "========= =end= ========\n\n"
			}
		}
	}

	return message
}

func sendToWeChat(message, token string) error {
	client := resty.New()
	wechatWebhookURL := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=" + token

	wechatMsg := WeChatMessage{
		MsgType: "markdown",
		Markdown: struct {
			Content string `json:"content"`
		}{
			Content: message,
		},
	}

	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(wechatMsg).
		Post(wechatWebhookURL)

	if err != nil {
		return fmt.Errorf("发送消息到企业微信失败: %v", err)
	}

	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("企业微信API返回错误状态码: %d, 响应: %s",
			response.StatusCode(), response.String())
	}

	return nil
}

func handleAlert(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	token := query.Get("token")
	env := query.Get("env")

	if token == "" {
		http.Error(w, "缺少token参数", http.StatusBadRequest)
		return
	}

	if env == "" {
		env = "测试"
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "读取请求体失败: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var alertRequest AlertRequest
	if err = json.Unmarshal(body, &alertRequest); err != nil {
		http.Error(w, "解析JSON失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 生成企业微信消息
	message := generateWeChatMessage(alertRequest, env)
	if message == "" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("没有需要发送的告警消息"))
		return
	}

	// 发送消息到企业微信
	if err = sendToWeChat(message, token); err != nil {
		http.Error(w, "发送消息到企业微信失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("告警消息已成功发送到企业微信"))
}

func main() {
	http.HandleFunc("/alert", handleAlert)
	port := "5050"
	log.Printf("Prometheus告警转发服务启动，监听端口: %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
