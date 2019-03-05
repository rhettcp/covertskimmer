package covertskimmer

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type CovertClient struct {
	httpClient  *http.Client
	username    string
	password    string
	lastLogin   *time.Time
	PlanDetails *PlanDetails `json:"PlanDetails"`
	CameraList  []*Camera    `json:"CameraList"`
}

type Camera struct {
	ID             string `json:"ID"`
	BatteryPercent string `json:"BatteryPercent"`
	SdCardSpace    string `json:"SdCardSpace"`
}

type PlanDetails struct {
	PlanTotal     string
	CurrentlyUsed string
}

func NewCovertClient(username string, password string) (*CovertClient, error) {
	if username == "" || password == "" {
		return nil, errors.New("Invalid Auth")
	}
	c := CovertClient{}
	cookieJar, _ := cookiejar.New(nil)
	c.httpClient = &http.Client{
		Jar: cookieJar,
	}
	c.username = username
	c.password = password
	c.PlanDetails = &PlanDetails{}
	err := c.login()
	if err != nil {
		return nil, err
	}
	err = c.findCameras()
	if err != nil {
		return nil, err
	}
	err = c.loadCameraStats()
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *CovertClient) GetImageList(camera Camera) ([]string, error) {
	if c.lastLogin == nil || c.lastLogin.Before(time.Now().Add(-12*time.Hour)) {
		err := c.login()
		if err != nil {
			return nil, err
		}
	}
	cameraURL := fmt.Sprintf("https://covert-wireless.com/photos?camera=%s", camera.GetID())

	req, err := http.NewRequest(http.MethodGet, cameraURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	pageData, _ := ioutil.ReadAll(resp.Body)
	pageString := string(pageData)
	links := make([]string, 0)
	for {
		index := strings.Index(pageString, "https://covert-camera-images.s3.amazonaws.com/")
		if index == -1 {
			break
		}
		pageString = pageString[index:]
		index2 := strings.Index(pageString, "\"")
		if index2 == -1 {
			break
		}
		links = append(links, strings.Replace(pageString[:index2], "/320/", "/1024/", 1))
		pageString = pageString[index2:]
	}
	return links, nil
}

func (c *CovertClient) findCameras() error {
	req, err := http.NewRequest(http.MethodGet, "https://covert-wireless.com/", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	pageData, _ := ioutil.ReadAll(resp.Body)
	pageString := string(pageData)
	cameras := make([]*Camera, 0)
	for {
		index := strings.Index(pageString, "/cameras/show?camera=")
		if index == -1 {
			break
		}
		pageString = pageString[index:]
		index2 := strings.Index(pageString, "#")
		if index2 == -1 {
			break
		}
		cameraID := strings.TrimPrefix(pageString[:index2], "/cameras/show?camera=")
		contained := false
		for _, b := range cameras {
			if b.ID == cameraID {
				contained = true
			}
		}
		if !contained {
			cam := Camera{ID: cameraID}
			cameras = append(cameras, &cam)
		}
		pageString = pageString[index2:]
	}
	c.CameraList = cameras
	return nil
}

func (c *CovertClient) login() error {
	form := url.Values{}
	form.Add("email", c.username)
	form.Add("password", c.password)

	req, err := http.NewRequest(http.MethodPost, "https://covert-wireless.com/login", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Invalid Response Code %d", resp.StatusCode)
	}
	t := time.Now()
	c.lastLogin = &t
	return nil
}

func (c *CovertClient) loadCameraStats() error {
	for _, cam := range c.CameraList {
		url := fmt.Sprintf("https://covert-wireless.com/cameras/show?camera=%s#info", cam.GetID())
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		pageData, _ := ioutil.ReadAll(resp.Body)
		pageString := string(pageData)
		pageString = trimTo(pageString, "<div class=\"cam-stat-left fll\">Available SD Card Space:</div>")
		pageString = trimTo(pageString, "<div class=\"cam-stat-val ovh\">")
		ind := strings.Index(pageString, "<")
		sd := pageString[:ind]
		cam.SdCardSpace = sd
		pageString = trimTo(pageString, "<div class=\"cam-stat-left fll\">Battery Level:</div>")
		pageString = trimTo(pageString, "<div class=\"cam-stat-val ovh\">")
		ind = strings.Index(pageString, "<")
		bat := pageString[:ind]
		cam.BatteryPercent = bat
		if c.PlanDetails.CurrentlyUsed != "" {
			continue
		}
		pageString = trimTo(pageString, "<div class=\"cam-stat-left fll\">Billing plan name:</div>")
		pageString = trimTo(pageString, "<div class=\"cam-stat-val ovh\">")
		ind = strings.Index(pageString, "<")
		planTotal := pageString[:ind]
		c.PlanDetails.PlanTotal = planTotal
		pageString = trimTo(pageString, "<div class=\"cam-stat-left fll\">Total photos:</div>")
		pageString = trimTo(pageString, "<div class=\"cam-stat-val ovh\">")
		ind = strings.Index(pageString, "<")
		used := pageString[:ind]
		c.PlanDetails.CurrentlyUsed = used
	}
	return nil
}

func (c *CovertClient) GetCameras() []*Camera {
	return c.CameraList
}

func (c *Camera) GetID() string {
	return c.ID
}

func (c *Camera) GetBattery() string {
	return c.BatteryPercent
}

func (c *Camera) GetSDCardSpace() string {
	return c.SdCardSpace
}

func trimTo(data, identifier string) string {
	ind := strings.Index(data, identifier)
	if ind == -1 || ind+len(identifier) > len(data) {
		return data
	}
	return data[ind+len(identifier):]
}
