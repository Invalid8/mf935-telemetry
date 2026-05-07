package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/invalid8/mf935-telemetry/internal/client"
)

type Session struct {
	client         *client.RouterClient
	waInnerVersion string
	crVersion      string
}

func New(c *client.RouterClient) *Session {
	return &Session{client: c}
}

func (s *Session) Login(password string) error {
	return s.LoginPrehashed(sha256Upper(password))
}

func (s *Session) LoginPrehashed(preHashed string) error {
	data, err := s.client.GetCmds([]string{"LD"})
	if err != nil {
		return fmt.Errorf("login: fetch LD: %w", err)
	}

	ld, ok := data["LD"]
	if !ok || ld == "" {
		return fmt.Errorf("login: LD token missing from response")
	}

	hashed := sha256Upper(preHashed + ld)

	result, err := s.client.Post(map[string]string{
		"goformId": "LOGIN",
		"password": hashed,
	})
	if err != nil {
		return fmt.Errorf("login: post: %w", err)
	}

	switch result["result"] {
	case "0", "4":
	case "1":
		return fmt.Errorf("login: incorrect password")
	case "2":
		return fmt.Errorf("login: duplicate session — already logged in")
	case "3":
		return fmt.Errorf("login: incorrect password")
	case "5":
		return fmt.Errorf("login: account locked")
	default:
		return fmt.Errorf("login: unexpected result: %s", result["result"])
	}

	if err := s.fetchVersions(); err != nil {
		return fmt.Errorf("login: fetch versions: %w", err)
	}

	return nil
}

func (s *Session) ADToken() (string, error) {
	if s.waInnerVersion == "" || s.crVersion == "" {
		return "", fmt.Errorf("ADToken: session not initialised — call Login first")
	}

	data, err := s.client.GetCmds([]string{"RD"})
	if err != nil {
		return "", fmt.Errorf("ADToken: fetch RD: %w", err)
	}

	rd, ok := data["RD"]
	if !ok || rd == "" {
		return "", fmt.Errorf("ADToken: RD missing from response")
	}

	step1 := sha256Upper(s.waInnerVersion + s.crVersion)
	ad := sha256Upper(step1 + rd)

	return ad, nil
}

func (s *Session) IsLoggedIn() (bool, error) {
	data, err := s.client.GetCmds([]string{"loginfo"})
	if err != nil {
		return false, fmt.Errorf("IsLoggedIn: %w", err)
	}
	return data["loginfo"] == "ok", nil
}

func (s *Session) fetchVersions() error {
	data, err := s.client.GetCmds([]string{"Language", "cr_version", "wa_inner_version"})
	if err != nil {
		return err
	}

	s.waInnerVersion = data["wa_inner_version"]
	s.crVersion = data["cr_version"]

	if s.waInnerVersion == "" || s.crVersion == "" {
		return fmt.Errorf("fetchVersions: one or both version strings are empty")
	}

	return nil
}

func sha256Upper(input string) string {
	sum := sha256.Sum256([]byte(input))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}
