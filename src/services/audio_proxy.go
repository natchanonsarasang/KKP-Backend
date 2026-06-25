package services

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type audioProxyService struct {
	client *resty.Client
}

// IAudioProxyService fetches an upstream audio file on behalf of the frontend,
// replicating the Supabase Edge Function this replaces (see src/gateways/audio_proxy.go).
type IAudioProxyService interface {
	FetchAudio(url string) (body []byte, contentType string, err error)
}

func NewAudioProxyService() IAudioProxyService {
	return &audioProxyService{client: resty.New()}
}

// AudioFetchError indicates the upstream audio URL could not be retrieved
// after both attempts (with spoofed headers, then a plain GET).
type AudioFetchError struct {
	StatusCode int
}

func (e *AudioFetchError) Error() string {
	return fmt.Sprintf("audio source is not accessible (status %d)", e.StatusCode)
}

func (sv *audioProxyService) FetchAudio(url string) ([]byte, string, error) {
	resp, err := sv.client.R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
		SetHeader("Referer", "https://voicebot.botnoi.ai/").
		SetHeader("Accept", "*/*").
		Get(url)

	if err != nil || resp.IsError() {
		// Retry once with a plain GET and no extra headers.
		resp, err = sv.client.R().Get(url)
		if err != nil || resp.IsError() {
			status := 0
			if resp != nil {
				status = resp.StatusCode()
			}
			return nil, "", &AudioFetchError{StatusCode: status}
		}
	}

	return resp.Body(), resp.Header().Get("Content-Type"), nil
}
