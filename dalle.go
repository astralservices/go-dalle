package dalle

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// Sizes
const (
	Small  int = 256
	Medium int = 512
	Large  int = 1024
)

const (
	defaultBaseURL   = "https://api.openai.com/v1/images"
	defaultUserAgent = "go-dalle"
	defaultTimeout   = 30 * time.Second
)

type Response struct {
	Created int64   `json:"created"`
	Data    []Datum `json:"data"`
}

type Datum struct {
	URL string `json:"url"`
}

const (
	URLFormat        = "url"
	Base64JSONFormat = "b64_json"
)

type GenerateRequest struct {
	Prompt         string  `json:"prompt"`
	N              *int    `json:"n,omitempty"`
	Size           *string `json:"size,omitempty"`
	ResponseFormat *string `json:"response_format,omitempty"`
	User           *string `json:"user,omitempty"`
}

type Client interface {
	Generate(prompt string, size *int, n *int, user *string, responseType *string) ([]Datum, error)
	Edit(prompt string, image *os.File, mask *os.File, size *int, n *int, user *string, responseType *string) ([]Datum, error)
	Variation(image *os.File, size *int, n *int, user *string, responseType *string) ([]Datum, error)
}

type client struct {
	baseURL    string
	apiKey     string
	userAgent  string
	httpClient *http.Client
}

func NewClient(apiKey string) Client {
	httpClient := &http.Client{
		Timeout: defaultTimeout,
	}

	c := &client{
		baseURL:    defaultBaseURL,
		apiKey:     apiKey,
		userAgent:  defaultUserAgent,
		httpClient: httpClient,
	}

	return c
}

func pointerizeString(s string) *string {
	return &s
}

// Prompt is the prompt to generate an image from.
//
// Size is the size of the image to generate (Small, Medium, Large).
//
// N is the number of images to generate.
//
// https://beta.openai.com/docs/guides/images/usage
func (c *client) Generate(prompt string, size *int, n *int, user *string, responseType *string) ([]Datum, error) {
	url := c.baseURL + "/generations"

	var sizeStr *string

	if size != nil {
		sizeStr = pointerizeString(fmt.Sprintf("%dx%d", size, size))
	}

	body := GenerateRequest{
		Prompt:         prompt,
		N:              n,
		Size:           sizeStr,
		User:           user,
		ResponseFormat: responseType,
	}

	jsonStr, err := json.Marshal(body)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case 400:
			return nil, errors.New("bad request")
		case 401:
			return nil, errors.New("unauthorized")
		case 403:
			return nil, errors.New("forbidden")
		case 404:
			return nil, errors.New("not found")
		case 429:
			return nil, errors.New("too many requests")
		case 500:
			return nil, errors.New("internal server error")
		case 502:
			return nil, errors.New("bad gateway")
		case 503:
			return nil, errors.New("service unavailable")
		case 504:
			return nil, errors.New("gateway timeout")
		default:
			return nil, errors.New("unknown error")
		}
	}

	var response Response

	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// Prompt is the prompt to generate an image from.
//
// Image is the image to edit.
//
// Mask is the mask to edit the image with.
//
// Size is the size of the image to generate (Small, Medium, Large).
//
// N is the number of images to generate.
//
// https://beta.openai.com/docs/guides/images/edits
func (c *client) Edit(prompt string, image *os.File, mask *os.File, size *int, n *int, user *string, responseType *string) ([]Datum, error) {
	url := c.baseURL + "/edits"

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	if image == nil {
		return nil, errors.New("image is nil")
	}

	if mask == nil {
		return nil, errors.New("mask is nil")
	}

	if imageWriter, err := w.CreateFormFile("image", image.Name()); err != nil {
		return nil, err
	} else if _, err := io.Copy(imageWriter, image); err != nil {
		return nil, err
	}

	if maskWriter, err := w.CreateFormFile("mask", mask.Name()); err != nil {
		return nil, err
	} else if _, err := io.Copy(maskWriter, mask); err != nil {
		return nil, err
	}

	var sizeStr *string

	if size != nil {
		sizeStr = pointerizeString(fmt.Sprintf("%dx%d", size, size))
	}

	err := w.WriteField("prompt", prompt)

	if err != nil {
		return nil, err
	}

	if n != nil {
		err = w.WriteField("n", fmt.Sprintf("%d", n))

		if err != nil {
			return nil, err
		}
	}

	if sizeStr != nil {
		err = w.WriteField("size", *sizeStr)

		if err != nil {
			return nil, err
		}
	}

	if user != nil {
		err = w.WriteField("user", *user)

		if err != nil {
			return nil, err
		}
	}

	if responseType != nil {
		err = w.WriteField("response_format", *responseType)

		if err != nil {
			return nil, err
		}
	}

	err = w.Close()

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, &b)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case 400:
			return nil, errors.New("bad request")
		case 401:
			return nil, errors.New("unauthorized")
		case 403:
			return nil, errors.New("forbidden")
		case 404:
			return nil, errors.New("not found")
		case 429:
			return nil, errors.New("too many requests")
		case 500:
			return nil, errors.New("internal server error")
		case 502:
			return nil, errors.New("bad gateway")
		case 503:
			return nil, errors.New("service unavailable")
		case 504:
			return nil, errors.New("gateway timeout")
		default:
			return nil, errors.New("unknown error")
		}
	}

	var response Response

	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// Image is the image to edit.
//
// Size is the size of the image to generate (Small, Medium, Large).
//
// N is the number of images to generate.
//
// https://beta.openai.com/docs/guides/images/variations
func (c *client) Variation(image *os.File, size *int, n *int, user *string, responseType *string) ([]Datum, error) {
	url := c.baseURL + "/variations"

	// this is posting using multipart/form-data

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	if image == nil {
		return nil, errors.New("image is nil")
	}

	imageWriter, err := w.CreateFormFile("image", image.Name())

	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(imageWriter, image); err != nil {
		return nil, err
	}

	var sizeStr *string

	if size != nil {
		sizeStr = pointerizeString(fmt.Sprintf("%dx%d", size, size))
	}

	if n != nil {
		err = w.WriteField("n", fmt.Sprintf("%d", n))

		if err != nil {
			return nil, err
		}
	}

	if sizeStr != nil {
		err = w.WriteField("size", *sizeStr)

		if err != nil {
			return nil, err
		}
	}

	if user != nil {
		err = w.WriteField("user", *user)

		if err != nil {
			return nil, err
		}
	}

	if responseType != nil {
		err = w.WriteField("response_format", *responseType)

		if err != nil {
			return nil, err
		}
	}

	err = w.Close()

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, &b)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case 400:
			return nil, errors.New("bad request")
		case 401:
			return nil, errors.New("unauthorized")
		case 403:
			return nil, errors.New("forbidden")
		case 404:
			return nil, errors.New("not found")
		case 429:
			return nil, errors.New("too many requests")
		case 500:
			return nil, errors.New("internal server error")
		case 502:
			return nil, errors.New("bad gateway")
		case 503:
			return nil, errors.New("service unavailable")
		case 504:
			return nil, errors.New("gateway timeout")
		default:
			return nil, errors.New("unknown error")
		}
	}

	var response Response

	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return nil, err
	}

	return response.Data, nil
}
