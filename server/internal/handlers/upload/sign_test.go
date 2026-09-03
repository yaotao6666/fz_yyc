package upload

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"fz_yyc_api/internal/config"
	"fz_yyc_api/pkg/qiniu"

	"github.com/gin-gonic/gin"
)

// setupSignRouter 构造仅含 POST /sign 路由的测试引擎，并初始化带密钥的七牛服务。
func setupSignRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewUploadHandler()
	r.POST("/sign", h.Sign)
	return r
}

// setupQiniuWithKeys 向全局注入一份可签名的七牛服务（UploadURL 非空以避免测试发起网络请求）
func setupQiniuWithKeys(t *testing.T) {
	t.Helper()

	config.Config = &config.AppConfig{
		Qiniu: config.Qiniu{
			AccessKey: "test-access-key",
			SecretKey: "test-secret-key",
			Bucket:    "test-bucket",
			Domain:    "https://cdn.example.com",
			UploadURL: "https://up.example.com",
		},
	}
	// 重置全局服务，保证测试环境干净
	if err := qiniu.InitQiniu(); err != nil {
		t.Fatalf("InitQiniu() failed: %v", err)
	}
}

// setupQiniuWithoutKeys 向全局注入一份未配置密钥的空七牛服务（对应生产未配置场景）
func setupQiniuWithoutKeys(t *testing.T) {
	t.Helper()

	config.Config = &config.AppConfig{
		Qiniu: config.Qiniu{},
	}
	if err := qiniu.InitQiniu(); err != nil {
		t.Fatalf("InitQiniu() failed: %v", err)
	}
}

func TestSignRoute(t *testing.T) {
	type respBody struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	do := func(t *testing.T, r *gin.Engine, body string) (*httptest.ResponseRecorder, respBody) {
		t.Helper()
		req := httptest.NewRequest("POST", "/sign", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var br respBody
		if err := json.Unmarshal(w.Body.Bytes(), &br); err != nil {
			t.Fatalf("响应不是有效 JSON: %v, body=%s", err, w.Body.String())
		}
		return w, br
	}

	t.Run("成功的私有签名", func(t *testing.T) {
		setupQiniuWithKeys(t)
		r := setupSignRouter(t)

		w, br := do(t, r, `{"url":"https://cdn.example.com/uploads/common/1.jpg"}`)
		if w.Code != 200 {
			t.Fatalf("HTTP 状态码 = %d, 期望 200", w.Code)
		}
		if br.Code != 0 {
			t.Fatalf("code = %d, 期望 0", br.Code)
		}
		if br.Data == nil || br.Data.URL == "" {
			t.Fatal("data.url 为空")
		}
		url := br.Data.URL
		if !strings.Contains(url, "token=") || !strings.Contains(url, "e=") {
			t.Fatalf("签名 URL 缺少 token/e 参数: %s", url)
		}
	})

	t.Run("非七牛域名透传不签名", func(t *testing.T) {
		setupQiniuWithKeys(t)
		r := setupSignRouter(t)

		_, br := do(t, r, `{"url":"https://other.com/x.png"}`)
		if br.Data == nil || br.Data.URL != "https://other.com/x.png" {
			t.Fatalf("外部域名应原样返回，got: %+v", br.Data)
		}
	})

	t.Run("七牛未配置密钥时不签名仅去query", func(t *testing.T) {
		setupQiniuWithoutKeys(t)
		r := setupSignRouter(t)

		_, br := do(t, r, `{"url":"https://cdn.example.com/uploads/common/1.jpg"}`)
		// 无密钥配置时 BuildPrivateURL 返回原 URL（不含 token），不应崩溃
		if br.Data == nil || strings.Contains(br.Data.URL, "token=") {
			t.Fatalf("未配置密钥时应返回原 URL，got: %+v", br.Data)
		}
	})

	t.Run("空 URL 返回参数错误", func(t *testing.T) {
		setupQiniuWithKeys(t)
		r := setupSignRouter(t)

		w, br := do(t, r, `{"url":""}`)
		if br.Code != 1001 {
			t.Fatalf("code = %d, 期望 1001", br.Code)
		}
		if w.Code != 400 {
			t.Fatalf("HTTP 状态码 = %d, 期望 400", w.Code)
		}
	})

	t.Run("缺失 URL 字段返回参数错误", func(t *testing.T) {
		setupQiniuWithKeys(t)
		r := setupSignRouter(t)

		_, br := do(t, r, `{}`)
		if br.Code != 1001 {
			t.Fatalf("code = %d, 期望 1001", br.Code)
		}
	})

	t.Run("非法 JSON 返回参数错误", func(t *testing.T) {
		setupQiniuWithKeys(t)
		r := setupSignRouter(t)

		_, br := do(t, r, `not-json`)
		if br.Code != 1001 {
			t.Fatalf("code = %d, 期望 1001", br.Code)
		}
	})
}