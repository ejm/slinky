package internal

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
)

const (
	DISCORD_INVITE_URL          = "https://discord.com/api/invites/%s?with_counts=true"
	DISCORD_DEFAULT_DESCRIPTION = "Join %s now!"
	DISCORD_TEMPLATE            = `
<meta charset="UTF-8">
<meta property="og:title" content="%s" />
<meta property="og:description" content="%s

🟢 %d Online ⚫ %d Members" />
<meta property="og:site_name" content="%s" />
<meta name="theme-color" content="%s" />`
)

type DiscordInvite struct {
	Code  string `json:"code"`
	Guild struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"guild"`
	MemberCount int `json:"approximate_member_count"`
	OnlineCount int `json:"approximate_presence_count"`
}

func (s *Server) writeDiscordInvite(invite *DiscordInvite, w http.ResponseWriter) error {
	w.Header().Add("Content-Type", "text/html")
	description := fmt.Sprintf(DISCORD_DEFAULT_DESCRIPTION, invite.Guild.Name)
	if invite.Guild.Description != "" {
		description = invite.Guild.Description
	}
	_, err := fmt.Fprintf(w, DISCORD_TEMPLATE,
		html.EscapeString(invite.Guild.Name),
		html.EscapeString(description),
		invite.OnlineCount,
		invite.MemberCount,
		html.EscapeString(s.config.Discord.PoweredBy),
		s.config.Discord.Color,
	)
	return err
}

// Requests a Discord invite for valid invite links
func (s *Server) extractDiscordInvite(link string) *DiscordInvite {
	inviteUrl, err := url.Parse(link)
	if err != nil {
		return nil
	}
	if inviteUrl.Host != "discord.gg" {
		return nil
	}
	if inviteUrl.Path == "" {
		return nil
	}
	endpoint := fmt.Sprintf(DISCORD_INVITE_URL, inviteUrl.Path)
	resp, err := s.httpClient.Get(endpoint)
	if err != nil {
		s.logger.Error(err.Error())
		return nil
	}
	defer resp.Body.Close()
	var invite DiscordInvite
	jsonParser := json.NewDecoder(resp.Body)
	err = jsonParser.Decode(&invite)
	if err != nil {
		s.logger.Error(err.Error())
		return nil
	}
	return &invite
}
