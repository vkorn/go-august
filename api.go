package august

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const (
	apiBase     = "https://api-production.august.com"
	tokenHeader = "x-august-access-token"
)

// ErrorGeneric describes a generic error.
type ErrorGeneric struct {
	Message string
}

// Error formats output.
func (e *ErrorGeneric) Error() string {
	return e.Message
}

// ErrorUnAuth describes a probable expired token error.
type ErrorUnAuth struct {
}

// Error formats output.
func (*ErrorUnAuth) Error() string {
	return "forbidden"
}

// Adds default headers.
func addHeaders(r *http.Request, token string) {
	r.Header.Add("Accept-Version", "0.0.1")
	r.Header.Add("x-august-api-key", "79fd0eb6-381d-4adf-95a0-47721289d1d9")
	r.Header.Add("x-kease-api-key", "79fd0eb6-381d-4adf-95a0-47721289d1d9")
	r.Header.Add("x-kease-api-key", "79fd0eb6-381d-4adf-95a0-47721289d1d9")
	r.Header.Add("Content-Type", "application/json")

	if "" != token {
		r.Header.Add(tokenHeader, token)
	}
}

// Performs August API call.
func doAPICall(a *Authenticator, method string, url string, reader io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, apiBase+url, reader)
	addHeaders(req, a.State.AccessToken)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
			return nil, &ErrorUnAuth{}
		}

		return nil, &ErrorGeneric{Message: "bad response"}
	}

	return resp, err
}

// APIProvider defines August APIs.
type APIProvider struct {
	auth *Authenticator
}

// NewAPIProvider constructs a new API provider.
func NewAPIProvider(auth *Authenticator) *APIProvider {
	p := &APIProvider{auth: auth}
	return p
}

// GetLocks returns all available locks.
func (p *APIProvider) GetLocks() ([]*Lock, error) {
	resp, err := doAPICall(p.auth, http.MethodGet, "/users/locks/mine", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	locksResp := make(map[string]*Lock)
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&locksResp)
	if err != nil {
		return nil, &ErrorGeneric{Message: err.Error()}
	}

	res := make([]*Lock, 0)

	for k, v := range locksResp {
		v.ID = k
		res = append(res, v)
	}

	return res, nil
}

// GetLockDetails returns detailed information about the lock.
func (p *APIProvider) GetLockDetails(lockID string) (*LockDetails, error) {
	resp, err := doAPICall(p.auth, http.MethodGet, "/locks/"+lockID, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	det := &LockDetails{}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(det)
	if err != nil {
		return nil, &ErrorGeneric{Message: err.Error()}
	}

	return det, nil
}

// GetLockStatus returns current lock's status.
func (p *APIProvider) GetLockStatus(lockID string) (*LockStatus, error) {
	resp, err := doAPICall(p.auth, http.MethodGet, "/locks/"+lockID+"/status", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	st := &LockStatus{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(st)
	if err != nil {
		return nil, &ErrorGeneric{Message: err.Error()}
	}

	return st, nil
}

type opTimeoutRequest struct {
	TimeoutSeconds int `json:"timeout"`
}

// Lock/unlock the lock.
func (p *APIProvider) lockOperation(lockID string, op string) error {
	req := &opTimeoutRequest{TimeoutSeconds: 30}
	b, err := json.Marshal(req)
	if err != nil {
		return &ErrorGeneric{Message: err.Error()}
	}
	_, err = doAPICall(p.auth, http.MethodPut, "/remoteoperate/"+lockID+"/"+op, bytes.NewReader(b))

	return err
}

// OpenLock unlocks the lock.
func (p *APIProvider) OpenLock(lockID string) error {
	return p.lockOperation(lockID, "unlock")
}

// CloseLock locks the lock.
func (p *APIProvider) CloseLock(lockID string) error {
	return p.lockOperation(lockID, "lock")
}
