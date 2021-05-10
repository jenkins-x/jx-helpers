package stringhelpers

import (
	"bytes"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// UrlJoin joins the given paths so that there is only ever one '/' character between the paths
func UrlJoin(paths ...string) string {
	var buffer bytes.Buffer
	last := len(paths) - 1
	query := ""
	for i, path := range paths {
		p, q := trimQuery(path)
		if q != "" {
			query = combineQuery(query, q)
		}
		if i > 0 {
			buffer.WriteString("/")
			p = strings.TrimPrefix(p, "/")
		}
		if i < last {
			p = strings.TrimSuffix(p, "/")
		}
		buffer.WriteString(p)
	}
	answer := buffer.String()
	if query == "" {
		return answer
	}
	return answer + "?" + query
}

func trimQuery(path string) (string, string) {
	idx := strings.Index(path, "?")
	if idx < 0 {
		return path, ""
	}
	return path[0:idx], path[idx+1:]
}

func combineQuery(q1 string, q2 string) string {
	if q1 == "" {
		return q2
	}
	if q2 == "" {
		return q1
	}
	return q1 + "&" + q2
}

// UrlHostNameWithoutPort returns the host name without any port of the given URL like string
func UrlHostNameWithoutPort(rawUri string) (string, error) {
	if strings.Index(rawUri, ":/") > 0 {
		u, err := url.Parse(rawUri)
		if err != nil {
			return "", err
		}
		rawUri = u.Host
	}

	// must be a crazy kind of string so lets do our best
	slice := strings.Split(rawUri, ":")
	idx := 0
	if len(slice) > 1 {
		if len(slice) > 2 {
			idx = 1
		}
		return strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(slice[idx], "/"), "/"), "/"), nil
	}
	return rawUri, nil
}

// UrlEqual verifies if URLs are equal
func UrlEqual(url1, url2 string) bool {
	return url1 == url2 || strings.TrimSuffix(url1, "/") == strings.TrimSuffix(url2, "/")
}

// SanitizeURL sanitizes by stripping the user and password
func SanitizeURL(unsanitizedUrl string) string {
	u, err := url.Parse(unsanitizedUrl)
	if err != nil {
		return unsanitizedUrl
	}
	return stripCredentialsFromURL(u)
}

// URLSetUserPassword adds the user and password to the URL
func URLSetUserPassword(rawURL, username, password string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, errors.Wrapf(err, "failed to parse URL %s", rawURL)
	}
	user := u.User
	if user != nil {
		if username == "" {
			username = user.Username()
		}
		if password == "" {
			password, _ = user.Password()
		}
	}
	u.User = url.UserPassword(username, password)
	return u.String(), nil
}

// stripCredentialsFromURL strip credentials from URL
func stripCredentialsFromURL(u *url.URL) string {
	pass, hasPassword := u.User.Password()
	userName := u.User.Username()
	if hasPassword {
		textToReplace := pass + "@"
		textToReplace = ":" + textToReplace
		if userName != "" {
			textToReplace = userName + textToReplace
		}
		return strings.Replace(u.String(), textToReplace, "", 1)
	}
	return u.String()
}

// URLToHostName converts the given URL to a host name returning the error string if its not a URL
func URLToHostName(svcURL string) string {
	host := ""
	if svcURL != "" {
		u, err := url.Parse(svcURL)
		if err != nil {
			host = err.Error()
		} else {
			host = u.Host
		}
	}
	return host
}

// IsValidUrl tests a string to determine if it is a well-structured url or not.
func IsValidUrl(s string) bool {
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}

	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}
