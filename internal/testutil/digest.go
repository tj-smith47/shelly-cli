package testutil

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// digestAlgorithm is the only digest algorithm Gen2+ devices use.
const digestAlgorithm = "SHA-256"

// digestQOP is the quality of protection Gen2+ digest responses use.
const digestQOP = "auth"

// DigestChallenge returns the WWW-Authenticate header value a Gen2+ Shelly
// device sends with a 401: HTTP digest, SHA-256, qop=auth.
func DigestChallenge(realm, nonce string) string {
	return fmt.Sprintf(`Digest qop="auth", realm=%q, nonce=%q, algorithm=SHA-256`, realm, nonce)
}

// DigestAuthorized reports whether r carries the Authorization header that
// answers DigestChallenge(realm, nonce) for the given user and password
// (RFC 7616, SHA-256, qop=auth). Fake devices use it to require a password
// the way a real Gen2+ device does.
func DigestAuthorized(r *http.Request, user, realm, nonce, password string) bool {
	return DigestAuthorizedHA1(r, user, realm, nonce, DigestHA1(user, realm, password))
}

// DigestHA1 returns the HA1 a Gen2+ device stores for user, realm and
// password: the hex SHA-256 of "user:realm:password".
func DigestHA1(user, realm, password string) string {
	return sha256Hex(user + ":" + realm + ":" + password)
}

// DigestAuthorizedHA1 is DigestAuthorized for a device that holds only the
// HA1 it was given through Shelly.SetAuth, as a real Gen2+ device does.
func DigestAuthorizedHA1(r *http.Request, user, realm, nonce, ha1 string) bool {
	header, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Digest ")
	if !ok {
		return false
	}
	params := map[string]string{}
	for part := range strings.SplitSeq(header, ",") {
		key, value, found := strings.Cut(strings.TrimSpace(part), "=")
		if found {
			params[key] = strings.Trim(value, `"`)
		}
	}
	if params["username"] != user || params["realm"] != realm ||
		params["nonce"] != nonce || params["uri"] != r.URL.RequestURI() ||
		params["qop"] != digestQOP || params["algorithm"] != digestAlgorithm {
		return false
	}

	ha2 := sha256Hex(r.Method + ":" + params["uri"])
	want := sha256Hex(strings.Join([]string{ha1, nonce, params["nc"], params["cnonce"], digestQOP, ha2}, ":"))
	return subtle.ConstantTimeCompare([]byte(params["response"]), []byte(want)) == 1
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// RPCDigestChallenge returns the message of the error 401 a Gen2+ device
// answers an unauthenticated request frame with: a JSON digest challenge.
// nonce keeps its JSON type: a string from firmware 2.0.0 on, a number
// before.
func RPCDigestChallenge(realm string, nonce any) string {
	data, err := json.Marshal(map[string]any{
		"auth_type": "digest", "nonce": nonce, "nc": 1, "realm": realm, "algorithm": digestAlgorithm,
	})
	if err != nil {
		panic(err)
	}
	return string(data)
}

// RPCDigestAuthorized reports whether auth, the auth object of a request
// frame, answers RPCDigestChallenge(realm, nonce) for user and password the
// way a Gen2+ device checks it: nonce echoed with its JSON type, cnonce a
// number, nc a string of 8 hex digits, HA2 = SHA256("dummy_method:dummy_uri")
// and response = SHA256(HA1:nonce:nc:cnonce:auth:HA2).
func RPCDigestAuthorized(auth json.RawMessage, user, realm string, nonce any, password string) bool {
	var a struct {
		Realm     string          `json:"realm"`
		Username  string          `json:"username"`
		Nonce     json.RawMessage `json:"nonce"`
		CNonce    json.RawMessage `json:"cnonce"`
		NC        string          `json:"nc"`
		Response  string          `json:"response"`
		Algorithm string          `json:"algorithm"`
	}
	if json.Unmarshal(auth, &a) != nil {
		return false
	}
	wantNonce, err := json.Marshal(nonce)
	if err != nil || string(a.Nonce) != string(wantNonce) {
		return false
	}
	var cnonce json.Number
	if json.Unmarshal(a.CNonce, &cnonce) != nil || len(a.CNonce) == 0 || a.CNonce[0] == '"' {
		return false
	}
	if a.Username != user || a.Realm != realm || a.Algorithm != digestAlgorithm || len(a.NC) != 8 {
		return false
	}

	ha1 := DigestHA1(user, realm, password)
	ha2 := sha256Hex("dummy_method:dummy_uri")
	want := sha256Hex(strings.Join([]string{ha1, fmt.Sprint(nonce), a.NC, cnonce.String(), digestQOP, ha2}, ":"))
	return subtle.ConstantTimeCompare([]byte(a.Response), []byte(want)) == 1
}
