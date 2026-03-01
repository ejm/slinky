package internal

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	sloghttp "github.com/samber/slog-http"
)

// General handler to send a 500 error and log it
func handleError(err error, w http.ResponseWriter, r *http.Request) {
	slog.ErrorContext(r.Context(), err.Error())
	http.Error(w, http.StatusText(500), http.StatusInternalServerError)
}

// Respond to a short link, either by redirecting or by generating a QR code
func (s *Server) shortLinkRedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortUrl := r.PathValue("link")
	shortUrl = strings.ToLower(shortUrl)
	// QR code handling if the link ends in `.png`
	makeQR := false
	if strings.HasSuffix(shortUrl, ".png") {
		shortUrl = strings.TrimSuffix(shortUrl, ".png")
		makeQR = true
	}
	sloghttp.AddCustomAttributes(r, slog.String("shortUrl", shortUrl))
	longUrl, err := s.resolveShortLink(shortUrl)
	if err != nil {
		handleError(err, w, r)
		return
	}
	if longUrl == nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		return
	}
	sloghttp.AddCustomAttributes(r, slog.String("longUrl", *longUrl))
	if makeQR {
		s.shortLinkQRCodeHandler(shortUrl, w, r)
		return
	}
	http.Redirect(w, r, *longUrl, http.StatusMovedPermanently)
}

func (s *Server) shortLinkQRCodeHandler(shortUrl string, w http.ResponseWriter, r *http.Request) {
	sizeStr := r.URL.Query().Get("size")
	size := 128
	if sizeStr != "" {
		if tmpSize, err := strconv.Atoi(sizeStr); err == nil {
			size = tmpSize
		}
	}

	shortUrl, err := url.JoinPath(s.config.BaseUrl, shortUrl)
	if err != nil {
		handleError(err, w, r)
		return
	}

	png, err := createQRCode(shortUrl, size)
	if err != nil {
		handleError(err, w, r)
		return
	}
	w.Header().Add("Content-Type", "image/png")
	w.Write(png)
}

type shortLinkRequest struct {
	ShortUrl *string `json:"vanity_url"`
	LongUrl  *string `json:"url"`
}

// Write JSON as an HTTP response
func writeJson(w http.ResponseWriter, i interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", " ")
	err := encoder.Encode(i)
	if err != nil {
		// TODO: Error as JSON
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) createShortLinkHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.InfoContext(r.Context(), "Creating short link")
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}
	var req shortLinkRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var resp *shortLinkResponse

	sloghttp.AddCustomAttributes(r, slog.String("longUrl", *req.LongUrl))

	if req.ShortUrl != nil {
		sloghttp.AddCustomAttributes(r, slog.String("shortUrl", *req.ShortUrl))
		resp, err = s.createVanityShortLink(*req.ShortUrl, *req.LongUrl)
	} else {
		resp, err = s.createShortLink(*req.LongUrl)
	}
	if err != nil {
		s.logger.ErrorContext(r.Context(), err.Error())
		if err == errAlreadyExists {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slog.InfoContext(
		r.Context(),
		"Created short link",
		"created", resp.Created,
		"longUrl", resp.LongUrl,
		"shortUrl", resp.ShortUrl,
		"vanity", resp.Vanity,
	)
	writeJson(w, resp)
}
