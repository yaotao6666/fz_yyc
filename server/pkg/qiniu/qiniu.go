package qiniu

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"fz_yyc_api/internal/config"
)

type QiniuService struct {
	accessKey string
	secretKey string
	bucket    string
	Domain    string
}

var qiniuService *QiniuService

type PutPolicy struct {
	Scope     string `json:"scope"`
	Expires   uint64 `json:"expires"`
	MimeLimit string `json:"mimeLimit,omitempty"`
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
		}
		return nil
	}
	qiniuService = &QiniuService{
		accessKey: cfg.AccessKey,
		secretKey: cfg.SecretKey,
		bucket:    cfg.Bucket,
		Domain:    cfg.Domain,
	}
	return nil
}

func GetService() *QiniuService {
	return qiniuService
}

func (q *QiniuService) GetUploadToken() (string, error) {
	if q.accessKey == "" || q.secretKey == "" || q.bucket == "" {
		return "", fmt.Errorf("七牛云未配置")
	}

	policy := PutPolicy{
		Scope:     q.bucket,
		Expires:   3600,
		MimeLimit: "image/*",
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("序列化策略失败: %v", err)
	}

	policyBase64 := base64.URLEncoding.EncodeToString(policyJSON)

	sign := hmac.New(sha1.New, []byte(q.secretKey))
	sign.Write([]byte(policyBase64))
	signature := fmt.Sprintf("%x", sign.Sum(nil))

	token := fmt.Sprintf("%s:%s:%s", q.accessKey, signature, policyBase64)
	return base64.URLEncoding.EncodeToString([]byte(token)), nil
}

func GenerateKey(prefix string, filename string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s/%d_%s", prefix, timestamp, filename)
}
