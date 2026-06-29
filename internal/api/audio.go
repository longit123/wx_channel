package api

import (
	"net/http"

	"wx_channel/internal/response"
	"wx_channel/internal/services"
)

// AudioAPI 处理音频提取相关请求
type AudioAPI struct {
	extractor *services.AudioExtractor
}

// NewAudioAPI 创建一个新的 AudioAPI
func NewAudioAPI(extractor *services.AudioExtractor) *AudioAPI {
	return &AudioAPI{extractor: extractor}
}

// GetCapabilities 返回当前环境的音频提取能力
// GET /api/v1/audio/capabilities
func (a *AudioAPI) GetCapabilities(w http.ResponseWriter, r *http.Request) {
	response.Success(w, a.extractor.Capabilities())
}

// RegisterRoutes 注册音频相关路由
func (a *AudioAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/audio/capabilities", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.ErrorWithStatus(w, http.StatusMethodNotAllowed, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		a.GetCapabilities(w, r)
	})
}
