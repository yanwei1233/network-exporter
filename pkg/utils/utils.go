package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	pconfig "github.com/prometheus/common/config"
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
	"time"
)

type RunMode string
type QuotaLevel string
type EventType string
type JobType string
type MetricsPhase string
type WarningEvent string

func TsToStr(input int64) string {
	return time.Unix(input, 0).Format("2006-01-02 15:04:05")
}
func Retry(attempts int, sleep time.Duration, opName string, fn func() error) error {
	if err := fn(); err != nil {
		if s, ok := err.(stop); ok {
			return s.error
		}
		if attempts--; attempts > 0 {
			//klog.Errorf("retry operation %s err: %s. attemps #%d after %s.", opName, err.Error(), attempts, sleep)
			time.Sleep(sleep)
			return Retry(attempts, 2*sleep, opName, fn)
		}
		return err
	}
	//klog.Infof("retry operation %s success", opName)
	return nil
}

type stop struct {
	error
}

func GetWithBearerToken(funcName string, hc pconfig.HTTPClientConfig, tw int, url string, params map[string]string) ([]byte, error) {
	//start := time.Now()
	client, err := pconfig.NewClientFromConfig(hc, funcName)
	if err != nil {
		klog.Errorf("[NewClientFromConfig.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}

	client.Timeout = time.Duration(tw) * time.Second
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		klog.Errorf("[GetWithBearerToken.http.NewRequest.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)

	}
	req.URL.RawQuery = q.Encode()
	klog.V(4).Infof("[NewClientFromConfig.result][funcName:%+v][url:%v][err:%v]", funcName, req.URL.String(), err)
	resp, err := client.Do(req)
	if err != nil {
		klog.Errorf("[GetWithBearerToken.request.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		klog.Errorf("[GetWithBearerToken.StatusCode.not2xx.error][funcName:%+v][url:%v][code:%v][resp_body_text:%v][err:%v]", funcName, url, resp.StatusCode,
			string(bodyBytes),
			err,
		)

		return nil, errors.New(fmt.Sprintf("server returned HTTP status %s", resp.Status))
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("[GetWithBearerToken.readbody.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	return bodyBytes, err
}

func GetWithBearerTokenArrayVars(funcName string, hc pconfig.HTTPClientConfig, tw int, url string, params map[string][]string) ([]byte, error) {
	//start := time.Now()
	client, err := pconfig.NewClientFromConfig(hc, funcName)
	if err != nil {
		klog.Errorf("[NewClientFromConfig.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}

	client.Timeout = time.Duration(tw) * time.Second
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		klog.Errorf("[GetWithBearerToken.http.NewRequest.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	q := req.URL.Query()
	for k, v := range params {
		for _, vv := range v {
			q.Add(k, vv)
		}

	}
	req.URL.RawQuery = q.Encode()
	klog.V(4).Infof("[NewClientFromConfig.result][funcName:%+v][url:%v][err:%v]", funcName, req.URL.String(), err)
	resp, err := client.Do(req)
	if err != nil {
		klog.Errorf("[GetWithBearerToken.request.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		klog.Errorf("[GetWithBearerToken.StatusCode.not2xx.error][funcName:%+v][url:%v][code:%v][resp_body_text:%v][err:%v]", funcName, url, resp.StatusCode,
			string(bodyBytes),
			err,
		)

		return nil, errors.New(fmt.Sprintf("server returned HTTP status %s", resp.Status))
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("[GetWithBearerToken.readbody.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	return bodyBytes, err
}

func PostWithApiToken(funcName string, apiToken string, tw int, url string, data interface{}) ([]byte, error) {
	//start := time.Now()
	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       time.Duration(tw) * time.Second,
	}

	bytesData, err := json.Marshal(data)
	if err != nil {
		klog.Errorf("[HttpPostPushDataJsonMarshalError][funcName:%v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	reader := bytes.NewReader(bytesData)

	req, err := http.NewRequest("POST", url, reader)
	token := fmt.Sprintf("ApiToken %s", apiToken)
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		klog.Errorf("[PostWithApiToken.http.NewRequest.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}

	klog.V(4).Infof("[NewClientFromConfig.result][funcName:%+v][url:%v][err:%v]", funcName, req.URL.String(), err)
	resp, err := client.Do(req)
	if err != nil {
		klog.Errorf("[PostWithApiToken.request.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		klog.Errorf("[PostWithApiToken.StatusCode.not2xx.error][funcName:%+v][url:%v][code:%v][resp_body_text:%v][err:%v]", funcName, url, resp.StatusCode,
			string(bodyBytes),
			err,
		)

		return nil, errors.New(fmt.Sprintf("server returned HTTP status %s", resp.Status))
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("[PostWithApiToken.readbody.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	return bodyBytes, err
}

func PostWithBearerToken(funcName string, hc pconfig.HTTPClientConfig, tw int, url string, data interface{}) ([]byte, error) {
	//start := time.Now()
	client, err := pconfig.NewClientFromConfig(hc, funcName)

	if err != nil {
		klog.Errorf("[NewClientFromConfig.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	bytesData, err := json.Marshal(data)
	if err != nil {
		klog.Errorf("[HttpPostPushDataJsonMarshalError][funcName:%v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	reader := bytes.NewReader(bytesData)

	client.Timeout = time.Duration(tw) * time.Second
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		klog.Errorf("[PostWithBearerToken.http.NewRequest.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}

	klog.V(4).Infof("[NewClientFromConfig.result][funcName:%+v][url:%v][err:%v]", funcName, req.URL.String(), err)
	resp, err := client.Do(req)
	if err != nil {
		klog.Errorf("[PostWithBearerToken.request.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		klog.Errorf("[PostWithBearerToken.StatusCode.not2xx.error][funcName:%+v][url:%v][code:%v][resp_body_text:%v][err:%v]", funcName, url, resp.StatusCode,
			string(bodyBytes),
			err,
		)

		return nil, errors.New(fmt.Sprintf("server returned HTTP status %s", resp.Status))
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("[PostWithBearerToken.readbody.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return nil, err
	}
	return bodyBytes, err
}

func DeleteWithBearerToken(funcName string, hc pconfig.HTTPClientConfig, tw int, url string) (bool, error) {
	//start := time.Now()
	client, err := pconfig.NewClientFromConfig(hc, funcName)
	if err != nil {
		klog.Errorf("[NewClientFromConfig.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return false, err
	}

	client.Timeout = time.Duration(tw) * time.Second
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		klog.Errorf("[DeleteWithBearerToken.http.NewRequest.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return false, err
	}
	klog.V(4).Infof("[NewClientFromConfig.result][funcName:%+v][url:%v][err:%v]", funcName, req.URL.String(), err)
	resp, err := client.Do(req)
	if err != nil {
		klog.Errorf("[DeleteWithBearerToken.request.error][funcName:%+v][url:%v][err:%v]", funcName, url, err)
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		klog.Errorf("[DeleteWithBearerToken.StatusCode.not2xx.error][funcName:%+v][url:%v][code:%v][resp_body_text:%v][err:%v]", funcName, url, resp.StatusCode,
			string(bodyBytes),
			err,
		)

		return false, errors.New(fmt.Sprintf("server returned HTTP status %s", resp.Status))
	}
	return true, nil
}

func FormatTenantProjectName(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	// remove more than one "-"
	l := len(s)
	if l <= 1 {
		return s
	}
	newS := "" + string(s[0])
	for i := 1; i < len(s); i++ {
		if s[i] == '-' && s[i] == s[i-1] {
			continue
		} else {
			if (s[i] <= 'z' && s[i] >= 'a') || (s[i] >= '0' && s[i] <= '9') || s[i] == '-' {
				newS = newS + string(s[i])
			} else {
				newS = newS + "-"
			}

		}
	}
	return newS
}

func SendMsgToIm(seatalkBotUrl string, content string, mentionedEmailList []string) {
	url := seatalkBotUrl
	type msgReq struct {
		Tag  string `json:"tag"`
		Text struct {
			Content            string   `json:"content"`
			MentionedEmailList []string `json:"mentioned_email_list"`
		} `json:"text"`
	}
	msg := msgReq{
		Tag: "text",
	}
	// msg  `[active][error] 3:37PM Name:DockerdDown Deploy: live Message:docker down on k8s-ci-rcmdplt-model-train node kube-ci-rcmdplt-model-train-10-131-148-231-node35
	//[View detail]: https://i.shp.ee/f4z924f
	msg.Text.Content = content
	msg.Text.MentionedEmailList = mentionedEmailList
	hc := pconfig.HTTPClientConfig{}
	res, err := PostWithBearerToken("SendMsgToIm", hc, 5, url, msg)
	if err != nil {
		klog.Errorf("[SendMsgToIm.http.NewRequest.error][url:%v][err:%v]", url, err)
		return
	}
	klog.Infof("[SendMsgToIm.result.print][url:%v][resp_body_text:%v]", url,
		string(res),
		err,
	)

}

func SendImageToIm(seatalkBotUrl string, imagePath string) {
	url := seatalkBotUrl
	type msgReq struct {
		Tag         string `json:"tag"`
		ImageBase64 struct {
			Content string `json:"content"`
		} `json:"image_base64"`
	}
	msg := msgReq{
		Tag: "image",
	}
	imageBytes, err := ioutil.ReadFile(imagePath)
	if err != nil {
		klog.Errorf("[SendImageToIm.LoadImageFile.error][url:%v][imagePath:%+v][err:%v]", url, imagePath, err)
		return
	}
	is := base64.StdEncoding.EncodeToString(imageBytes)

	msg.ImageBase64.Content = is
	hc := pconfig.HTTPClientConfig{}
	res, err := PostWithBearerToken("SendMsgToIm", hc, 5, url, msg)
	if err != nil {
		klog.Errorf("[SendImageToIm.http.NewRequest.error][url:%v][err:%v]", url, err)
		return
	}
	klog.Infof("[SendImageToIm.result.print][url:%v][resp_body_text:%v]", url,
		string(res),
		err,
	)

}
