package qiniu

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"fz_yyc_api/internal/config"
)

type QiniuService struct {
	accessKey string
	secretKey string
	bucket    string
	Domain    string
	UploadURL string
}

var qiniuService *QiniuService

type PutPolicy struct {
	Scope     string `json:"scope"`
	Deadline  uint64 `json:"deadline"`
	MimeLimit string `json:"mimeLimit,omitempty"`
}

type regionHosts struct {
	Main []string `json:"main"`
}

type regionGroup struct {
	Acc *regionHosts `json:"acc"`
	Src *regionHosts `json:"src"`
}

type bucketRegionResponse struct {
	Up *regionGroup `json:"up"`
}

func resolveUploadURL(accessKey string, bucket string, configuredURL string) string {
	if configuredURL != "" {
		return configuredURL
	}

	queryURL := fmt.Sprintf("https://uc.qbox.me/v2/query?ak=%s&bucket=%s", url.QueryEscape(accessKey), url.QueryEscape(bucket))
	resp, err := http.Get(queryURL)
	if err != nil {
		return "https://up.qiniup.com"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "https://up.qiniup.com"
	}

	var region bucketRegionResponse
	if err := json.NewDecoder(resp.Body).Decode(&region); err != nil {
		return "https://up.qiniup.com"
	}

	if region.Up != nil && region.Up.Src != nil && len(region.Up.Src.Main) > 0 && region.Up.Src.Main[0] != "" {
		return "https://" + region.Up.Src.Main[0]
	}

	if region.Up != nil && region.Up.Acc != nil && len(region.Up.Acc.Main) > 0 && region.Up.Acc.Main[0] != "" {
		return "https://" + region.Up.Acc.Main[0]
	}

	return "https://up.qiniup.com"
}

func InitQiniu() error {
	cfg := config.Config.Qiniu
	if cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Bucket == "" {
		log.Printf("七牛云配置不完整，跳过初始化")
		qiniuService = &QiniuService{
			accessKey: "",
			secretKey: "",
			bucket:    "",
			Domain:    "",
			UploadURL: "",
		}
		return nil
	}
	qiniuService = &QiniuService{
		accessKey: cfg.AccessKey,
		secretKey: cfg.SecretKey,
		bucket:    cfg.Bucket,
		Domain:    cfg.Domain,
		UploadURL: resolveUploadURL(cfg.AccessKey, cfg.Bucket, cfg.UploadURL),
	}
	return nil
}

func GetService() *QiniuService {
	return qiniuService
}

func (q *QiniuService) GetUploadToken() (string, error) {
	return q.GetUploadTokenWithMime("image/*")
}

// GetUploadTokenWithMime 生成限定 MIME 类型的上传凭证（如 audio/*）
func (q *QiniuService) GetUploadTokenWithMime(mimeLimit string) (string, error) {
	if q.accessKey == "" || q.secretKey == "" || q.bucket == "" {
		return "", fmt.Errorf("七牛云未配置")
	}

	policy := PutPolicy{
		Scope:     q.bucket,
		// 七牛要求 deadline 为绝对 Unix 时间戳，传相对秒数会被直接判定为已过期。
		Deadline:  uint64(time.Now().Add(time.Hour).Unix()),
		MimeLimit: mimeLimit,
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("序列化策略失败: %v", err)
	}

	policyBase64 := base64.URLEncoding.EncodeToString(policyJSON)

	sign := hmac.New(sha1.New, []byte(q.secretKey))
	sign.Write([]byte(policyBase64))
	signature := base64.URLEncoding.EncodeToString(sign.Sum(nil))

	return fmt.Sprintf("%s:%s:%s", q.accessKey, signature, policyBase64), nil
}

// DeleteFile 删除七牛空间内的文件（服务录音 30 天清理用）
// 入参 key 为对象键；若传入完整 URL 会自动剥离域名前缀
func (q *QiniuService) DeleteFile(key string) error {
	if q.accessKey == "" || q.secretKey == "" || q.bucket == "" {
		return fmt.Errorf("七牛云未配置")
	}

	key = extractObjectKey(key, q.Domain)
	if key == "" {
		return fmt.Errorf("文件key为空")
	}

	// 七牛删除接口：POST /delete/<encodedEntryURI>，QBox 签名
	entryURI := q.bucket + ":" + key
	encodedEntry := base64.URLEncoding.EncodeToString([]byte(entryURI))
	pathAndQuery := "/delete/" + encodedEntry

	sign := hmac.New(sha1.New, []byte(q.secretKey))
	sign.Write([]byte(pathAndQuery))
	signature := base64.URLEncoding.EncodeToString(sign.Sum(nil))
	accessToken := fmt.Sprintf("%s:%s", q.accessKey, signature)

	req, err := http.NewRequest("POST", "https://rs.qiniuapi.com"+pathAndQuery, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "QBox "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 200=删除成功 612=文件不存在（视为成功，幂等）
	if resp.StatusCode != http.StatusOK && resp.StatusCode != 612 {
		return fmt.Errorf("七牛删除失败: HTTP %d", resp.StatusCode)
	}
	return nil
}

// extractObjectKey 从完整 URL 或裸 key 中提取对象键
func extractObjectKey(resource string, domain string) string {
	resource = strings.TrimSpace(resource)
	if resource == "" {
		return ""
	}
	if domain != "" {
		domain = strings.TrimRight(domain, "/")
		if strings.HasPrefix(resource, domain+"/") {
			return strings.TrimPrefix(resource, domain+"/")
		}
	}
	if strings.HasPrefix(resource, "http://") || strings.HasPrefix(resource, "https://") {
		parsed, err := url.Parse(resource)
		if err != nil {
			return ""
		}
		return strings.TrimPrefix(parsed.Path, "/")
	}
	return strings.TrimLeft(resource, "/")
}

func (q *QiniuService) BuildPrivateURL(resource string) string {
	if resource == "" {
		return ""
	}

	baseURL, shouldSign := q.normalizeResourceURL(resource)
	if baseURL == "" || q.accessKey == "" || q.secretKey == "" {
		return baseURL
	}
	if !shouldSign {
		return baseURL
	}

	deadline := time.Now().Add(time.Hour).Unix()
	separator := "?"
	if strings.Contains(baseURL, "?") {
		separator = "&"
	}

	downloadURL := fmt.Sprintf("%s%se=%d", baseURL, separator, deadline)
	sign := hmac.New(sha1.New, []byte(q.secretKey))
	sign.Write([]byte(downloadURL))
	token := fmt.Sprintf("%s:%s", q.accessKey, base64.URLEncoding.EncodeToString(sign.Sum(nil)))
	return fmt.Sprintf("%s&token=%s", downloadURL, url.QueryEscape(token))
}

func (q *QiniuService) normalizeResourceURL(resource string) (string, bool) {
	if resource == "" {
		return "", false
	}

	resource = strings.TrimSpace(resource)
	if resource == "" {
		return "", false
	}

	if strings.HasPrefix(resource, "http://") || strings.HasPrefix(resource, "https://") {
		parsedResource, err := url.Parse(resource)
		if err != nil {
			return resource, false
		}

		parsedResource.RawQuery = ""
		parsedResource.Fragment = ""

		if q.Domain == "" {
			return parsedResource.String(), false
		}

		parsedDomain, err := url.Parse(q.Domain)
		if err != nil {
			return parsedResource.String(), false
		}

		if !strings.EqualFold(parsedResource.Host, parsedDomain.Host) {
			return parsedResource.String(), false
		}

		return parsedResource.String(), true
	}

	domain := strings.TrimRight(q.Domain, "/")
	path := strings.TrimLeft(resource, "/")
	if domain == "" {
		return path, false
	}

	return domain + "/" + path, true
}

// imgSrcRe 匹配 HTML 正文中的图片地址（含双引号/单引号，不区分大小写）。
// 分组：1=src="或src='，2=URL，3=闭合引号。
var imgSrcRe = regexp.MustCompile(`(?i)(src\s*=\s*["'])([^"']+)(["'])`)

// SignContentImages 对 HTML 正文中每个 <img src="..."> 的七牛私有地址进行签名。
// 非七牛域名或空资源会被 BuildPrivateURL 原样透传，安全无副作用。
func (q *QiniuService) SignContentImages(content string) string {
	if content == "" || q == nil {
		return content
	}
	return imgSrcRe.ReplaceAllStringFunc(content, func(m string) string {
		sub := imgSrcRe.FindStringSubmatch(m)
		if len(sub) != 4 {
			return m
		}
		return sub[1] + q.BuildPrivateURL(sub[2]) + sub[3]
	})
}

// StripContentSignatures 去除 HTML 正文中每个 <img src="..."> 的七牛签名参数（e、token），
// 还原为未签名的原始地址，用于入库持久化。
func (q *QiniuService) StripContentSignatures(content string) string {
	if content == "" {
		return content
	}
	return imgSrcRe.ReplaceAllStringFunc(content, func(m string) string {
		sub := imgSrcRe.FindStringSubmatch(m)
		if len(sub) != 4 {
			return m
		}
		return sub[1] + stripSignatureQuery(sub[2]) + sub[3]
	})
}

// stripSignatureQuery 从 URL 中移除七牛下载签名的 query 参数（e/token），保留其余参数。
func stripSignatureQuery(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := parsed.Query()
	query.Del("e")
	query.Del("token")
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func GenerateKey(prefix string, filename string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s/%d_%s", prefix, timestamp, filename)
}
