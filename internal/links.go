package internal

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Special case error for a row already existing
var errAlreadyExists = errors.New("short url already exists")

type shortLinkResponse struct {
	ShortUrl string    `json:"short_url"`
	LongUrl  string    `json:"url"`
	Vanity   bool      `json:"vanity"`
	Created  time.Time `json:"created"`
}

// Insert a link into the database
func (s *Server) insertLink(shortUrl string, longUrl string, vanity bool) (*shortLinkResponse, error) {
	rows, err := s.db.Query(
		"INSERT INTO links (short_url, long_url, vanity) VALUES ($1, $2, $3) RETURNING short_url, long_url, vanity, created;",
		strings.ToLower(shortUrl),
		longUrl,
		vanity,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			// SQLSTATE 23505 is the code for 'unique_violation'
			// We special case this because different types of links handle this error type differently
			if pqErr.Code == "23505" {
				return nil, errAlreadyExists
			}
		}
		return nil, err
	}
	defer rows.Close()

	var link shortLinkResponse
	for rows.Next() {
		err = rows.Scan(&link.ShortUrl, &link.LongUrl, &link.Vanity, &link.Created)
		if err != nil {
			return nil, err
		}
	}
	return &link, nil
}

// Create a vanity short link
func (s *Server) createVanityShortLink(shortUrl string, longUrl string) (*shortLinkResponse, error) {
	return s.insertLink(shortUrl, longUrl, true)
}

// Create a generated short link, retrying if the generated short URL exists
func (s *Server) createShortLink(longUrl string) (*shortLinkResponse, error) {
	for range s.config.MaxRetries {
		shortUrl, err := GenerateShortUrl(s.config.LinkSize)
		if err != nil {
			return nil, err
		}
		link, err := s.insertLink(shortUrl, longUrl, false)
		if err != nil {
			if err == errAlreadyExists {
				s.logger.Warn("Short link was already taken, re-rolling", "longUrl", longUrl, "shortUrl", shortUrl)
				continue
			}
			return nil, err
		}
		return link, nil
	}
	return nil, errors.New("maximum retries exceeded")
}

// Resolve a short url to a long url, return `nil` if the link doesn't exist
func (s *Server) resolveShortLink(shortUrl string) (*string, error) {
	var link string
	row := s.db.QueryRow("SELECT long_url FROM links WHERE short_url = $1;", shortUrl)
	err := row.Scan(&link)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		s.logger.Error(err.Error())
		return nil, err
	}
	return &link, nil
}
