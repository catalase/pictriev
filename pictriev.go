package pictriev

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
)

func dictate(command string, v url.Values) (*http.Response, error) {
	urlstr := "http://www.pictriev.com/facedbj.php?" + command + "&" + v.Encode()
	return http.Get(urlstr)
}

type FindFaceResult struct {
	ImageID string `json:"imageid"`
	// NFaces value is not 0 unless a faces cannot be found.
	NFaces int     `json:"nfaces"`
	PTime  float64 `json:"ptime"`
	Width  int     `json:"sx"`
	Height int     `json:"sy"`
}

type Attr struct {
	P       float64
	Name    string
	ImageID string
}

type Gender float64

func (u Gender) Man() float64 {
	return float64(u)
}

func (u Gender) Woman() float64 {
	return 1 - float64(u)
}

type WhoisResult struct {
	Age     float64   `json:"age"`
	AgeDist []float64 `json:"agedist"`
	// Attrs 는 높은 확률 순서로 정렬되어 있다.
	Attrs  []Attr
	Gender Gender
	Lang   Lang
}

type Fault struct {
	Command string
	Result  string `json:"result"`
	Msg     string `json:"msg"`
}

func (err *Fault) Error() string {
	if len(err.Msg) > 0 {
		return err.Msg
	}
	return err.Result
}

var errBadCode = errors.New("server does not respond propertly")

func parseFindFaceResult(resp *http.Response) (*FindFaceResult, error) {
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errBadCode
	}

	data := new(struct {
		FindFaceResult
		Fault
	})
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(data); err != nil {
		return nil, err
	}
	if data.Result != "OK" {
		data.Result = "FindFace"
		return nil, &data.Fault
	}

	return &data.FindFaceResult, nil
}

func FindFaceImageURL(urlstr string) (*FindFaceResult, error) {
	resp, err := dictate("findface", url.Values{
		"image": []string{urlstr},
	})
	if err != nil {
		return nil, err
	}

	return parseFindFaceResult(resp)
}

func FindFaceImage(src io.Reader) (*FindFaceResult, error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="photo"; filename="photo"`)

	part, err := w.CreatePart(h)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, src); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	resp, err := http.Post("http://www.pictriev.com/facedbj.php?findface&image=post",
		w.FormDataContentType(), buf)
	if err != nil {
		return nil, err
	}

	return parseFindFaceResult(resp)
}

type Lang string

const (
	De Lang = "de"
	En      = "en"
	Es      = "es"
	Fr      = "fr"
	Id      = "id"
	It      = "it"
	Ja      = "ja"
	Ko      = "ko"
	Pl      = "pl"
	Pt      = "pt"
	Ru      = "ru"
	Th      = "th"
	Tr      = "tr"
	Zh      = "zh"
)

// 아래 3 가2ㅣ에 대하여 추가적인 처리를 요구함
// 1. Attr
// 2. Gender
// 3. Result

func Whois(imageid string, faceid int, lang Lang) (*WhoisResult, error) {
	v := url.Values{}
	v.Set("imageid", imageid)
	v.Set("faceid", strconv.Itoa(faceid))
	v.Set("lang", string(lang))
	resp, err := dictate("whoissim", v)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errBadCode
	}

	data := new(struct {
		WhoisResult
		Fault
		RawAttrs  [][4]interface{} `json:"attrs"`
		RawGender [2]interface{}   `json:"gender"`
	})
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}

	if data.Result != "OK" {
		data.Command = "Whois"
		return nil, &data.Fault
	}

	data.Attrs = make([]Attr, len(data.RawAttrs))
	for i, raw := range data.RawAttrs {
		data.Attrs[i] = Attr{
			P:       raw[1].(float64),
			Name:    raw[2].(string),
			ImageID: strconv.FormatInt(int64(raw[3].(float64)), 10),
		}
	}

	gender := data.RawGender[0].(string)
	p := data.RawGender[1].(float64)
	if gender == "M" {
		data.Gender = Gender(p)
	} else {
		data.Gender = Gender(1 - p)
	}

	return &data.WhoisResult, nil
}
